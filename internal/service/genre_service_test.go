package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	"golang-project/internal/models"
)

type memoryGenreRepo struct {
	data map[int]*models.Genre
}

func newMemoryGenreRepo() *memoryGenreRepo {
	return &memoryGenreRepo{data: make(map[int]*models.Genre)}
}

func (r *memoryGenreRepo) GetAll(ctx context.Context) ([]models.Genre, error) {
	result := make([]models.Genre, 0, len(r.data))
	for _, g := range r.data {
		result = append(result, *g)
	}
	return result, nil
}

func (r *memoryGenreRepo) GetByID(ctx context.Context, id int) (*models.Genre, error) {
	if g, ok := r.data[id]; ok {
		return g, nil
	}
	return nil, sql.ErrNoRows
}

func (r *memoryGenreRepo) GetByName(ctx context.Context, name string) (*models.Genre, error) {
	for _, g := range r.data {
		if g.Name == name {
			return g, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memoryGenreRepo) Create(ctx context.Context, genre *models.Genre) error {
	now := time.Now()
	genre.ID = len(r.data) + 1
	genre.CreatedAt = now
	r.data[genre.ID] = genre
	return nil
}

func (r *memoryGenreRepo) Update(ctx context.Context, genre *models.Genre) error {
	if _, ok := r.data[genre.ID]; !ok {
		return sql.ErrNoRows
	}
	r.data[genre.ID] = genre
	return nil
}

func (r *memoryGenreRepo) Delete(ctx context.Context, id int) error {
	if _, ok := r.data[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.data, id)
	return nil
}

func TestGenreService_Create(t *testing.T) {
	repo := newMemoryGenreRepo()
	svc := NewGenreService(repo, validator.New())

	t.Run("ok", func(t *testing.T) {
		genre, err := svc.Create(context.Background(), models.CreateGenreRequest{Name: "Drama"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if genre.ID == 0 {
			t.Fatalf("expected ID set")
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		repo.data = make(map[int]*models.Genre)
		_ = repo.Create(context.Background(), &models.Genre{ID: 1, Name: "Comedy", CreatedAt: time.Now()})

		if _, err := svc.Create(context.Background(), models.CreateGenreRequest{Name: "Comedy"}); !errors.Is(err, ErrGenreExists) {
			t.Fatalf("expected ErrGenreExists, got %v", err)
		}
	})

	t.Run("validation", func(t *testing.T) {
		if _, err := svc.Create(context.Background(), models.CreateGenreRequest{Name: ""}); err == nil {
			t.Fatalf("expected validation error")
		}
	})
}

func TestGenreService_Get_Update_Delete(t *testing.T) {
	repo := newMemoryGenreRepo()
	svc := NewGenreService(repo, validator.New())

	g := &models.Genre{ID: 1, Name: "Action", CreatedAt: time.Now()}
	repo.data[g.ID] = g

	t.Run("get ok", func(t *testing.T) {
		got, err := svc.Get(context.Background(), g.ID)
		if err != nil || got.ID != g.ID {
			t.Fatalf("expected genre, got %v err %v", got, err)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		if _, err := svc.Get(context.Background(), 9999); !errors.Is(err, ErrGenreNotFound) {
			t.Fatalf("expected ErrGenreNotFound, got %v", err)
		}
	})

	t.Run("update ok", func(t *testing.T) {
		req := models.CreateGenreRequest{Name: "Adventure"}
		updated, err := svc.Update(context.Background(), g.ID, req)
		if err != nil || updated.Name != "Adventure" {
			t.Fatalf("expected update, got %v err %v", updated, err)
		}
	})

	t.Run("delete ok", func(t *testing.T) {
		if err := svc.Delete(context.Background(), g.ID); err != nil {
			t.Fatalf("expected delete ok, got %v", err)
		}
	})

	t.Run("delete not found", func(t *testing.T) {
		if err := svc.Delete(context.Background(), 9999); !errors.Is(err, ErrGenreNotFound) {
			t.Fatalf("expected ErrGenreNotFound, got %v", err)
		}
	})
}

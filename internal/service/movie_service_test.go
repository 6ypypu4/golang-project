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

type memoryMovieRepo struct {
	movies      map[int]*models.Movie
	movieGenres map[int][]int
}

type movieTestGenreRepo struct {
	data map[int]*models.Genre
}

func (g *movieTestGenreRepo) GetByID(ctx context.Context, id int) (*models.Genre, error) {
	if genre, ok := g.data[id]; ok {
		return genre, nil
	}
	return nil, sql.ErrNoRows
}

func newMemoryMovieRepo() *memoryMovieRepo {
	return &memoryMovieRepo{
		movies:      make(map[int]*models.Movie),
		movieGenres: make(map[int][]int),
	}
}

func (r *memoryMovieRepo) List(ctx context.Context, filters models.MovieFilters, limit, offset int) ([]models.Movie, int, error) {
	result := make([]models.Movie, 0, len(r.movies))
	for _, m := range r.movies {
		result = append(result, *m)
	}
	return result, len(result), nil
}

func (r *memoryMovieRepo) GetByID(ctx context.Context, id int) (*models.Movie, error) {
	if m, ok := r.movies[id]; ok {
		return m, nil
	}
	return nil, sql.ErrNoRows
}

func (r *memoryMovieRepo) Create(ctx context.Context, movie *models.Movie) error {
	now := time.Now()
	movie.ID = len(r.movies) + 1
	movie.CreatedAt = now
	movie.UpdatedAt = now
	r.movies[movie.ID] = movie
	return nil
}

func (r *memoryMovieRepo) Update(ctx context.Context, movie *models.Movie) error {
	if _, ok := r.movies[movie.ID]; !ok {
		return sql.ErrNoRows
	}
	movie.UpdatedAt = time.Now()
	r.movies[movie.ID] = movie
	return nil
}

func (r *memoryMovieRepo) Delete(ctx context.Context, id int) error {
	if _, ok := r.movies[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.movies, id)
	delete(r.movieGenres, id)
	return nil
}

func (r *memoryMovieRepo) SetGenres(ctx context.Context, movieID int, genreIDs []int) error {
	if _, ok := r.movies[movieID]; !ok {
		return sql.ErrNoRows
	}
	r.movieGenres[movieID] = genreIDs
	return nil
}

func (r *memoryMovieRepo) GetGenresByMovieID(ctx context.Context, movieID int) ([]models.Genre, error) {
	ids := r.movieGenres[movieID]
	result := make([]models.Genre, 0, len(ids))
	for _, id := range ids {
		result = append(result, models.Genre{ID: id, Name: ""})
	}
	return result, nil
}

func TestMovieService_Create(t *testing.T) {
	movieRepo := newMemoryMovieRepo()
	genreID := 1
	genreLookup := &movieTestGenreRepo{data: map[int]*models.Genre{
		genreID: {ID: genreID, Name: "Drama"},
	}}
	svc := NewMovieService(movieRepo, genreLookup, validator.New())

	t.Run("ok", func(t *testing.T) {
		req := models.CreateMovieRequest{
			Title:           "Inception",
			Description:     "Dream",
			ReleaseYear:     2010,
			Director:        "Nolan",
			DurationMinutes: 120,
			GenreIDs:        []string{"1"},
		}
		movie, err := svc.Create(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if movie.ID == 0 {
			t.Fatalf("expected ID set")
		}
		if len(movie.Genres) != 1 || movie.Genres[0].ID != genreID {
			t.Fatalf("expected genre assigned")
		}
	})

	t.Run("missing genres", func(t *testing.T) {
		req := models.CreateMovieRequest{
			Title:           "No genre",
			ReleaseYear:     2010,
			DurationMinutes: 100,
		}
		if _, err := svc.Create(context.Background(), req); err == nil {
			t.Fatalf("expected error for missing genres")
		}
	})

	t.Run("genre not found", func(t *testing.T) {
		req := models.CreateMovieRequest{
			Title:           "Bad genre",
			ReleaseYear:     2010,
			DurationMinutes: 100,
			GenreIDs:        []string{"9999"},
		}
		if _, err := svc.Create(context.Background(), req); !errors.Is(err, ErrGenreNotFound) {
			t.Fatalf("expected ErrGenreNotFound, got %v", err)
		}
	})
}

func TestMovieService_Update_Delete(t *testing.T) {
	movieRepo := newMemoryMovieRepo()
	genreID := 1
	genreLookup := &movieTestGenreRepo{data: map[int]*models.Genre{
		genreID: {ID: genreID, Name: "Drama"},
	}}
	svc := NewMovieService(movieRepo, genreLookup, validator.New())

	// seed movie
	m := &models.Movie{
		ID:              1,
		Title:           "Old",
		ReleaseYear:     2000,
		DurationMinutes: 90,
		Director:        "Dir",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	movieRepo.movies[m.ID] = m
	movieRepo.movieGenres[m.ID] = []int{genreID}

	t.Run("update ok", func(t *testing.T) {
		req := models.UpdateMovieRequest{
			Title:       "New title",
			Description: "Desc",
			GenreIDs:    []string{"1"},
		}
		updated, err := svc.Update(context.Background(), m.ID, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Title != "New title" || updated.Description != "Desc" {
			t.Fatalf("expected fields updated")
		}
	})

	t.Run("update not found", func(t *testing.T) {
		req := models.UpdateMovieRequest{Title: "X"}
		if _, err := svc.Update(context.Background(), 9999, req); !errors.Is(err, ErrMovieNotFound) {
			t.Fatalf("expected ErrMovieNotFound, got %v", err)
		}
	})

	t.Run("delete ok", func(t *testing.T) {
		if err := svc.Delete(context.Background(), m.ID); err != nil {
			t.Fatalf("expected delete ok, got %v", err)
		}
	})

	t.Run("delete not found", func(t *testing.T) {
		if err := svc.Delete(context.Background(), 9999); !errors.Is(err, ErrMovieNotFound) {
			t.Fatalf("expected ErrMovieNotFound, got %v", err)
		}
	})
}

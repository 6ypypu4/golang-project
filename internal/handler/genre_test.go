package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"golang-project/internal/models"
	"golang-project/internal/service"
)

// in-memory repo for handler tests
type ghRepo struct {
	data map[int]*models.Genre
}

func newGHRepo() *ghRepo {
	return &ghRepo{data: make(map[int]*models.Genre)}
}

func (r *ghRepo) GetAll(ctx context.Context) ([]models.Genre, error) {
	result := make([]models.Genre, 0, len(r.data))
	for _, g := range r.data {
		result = append(result, *g)
	}
	return result, nil
}

func (r *ghRepo) GetByID(ctx context.Context, id int) (*models.Genre, error) {
	if g, ok := r.data[id]; ok {
		return g, nil
	}
	return nil, sql.ErrNoRows
}

func (r *ghRepo) GetByName(ctx context.Context, name string) (*models.Genre, error) {
	for _, g := range r.data {
		if g.Name == name {
			return g, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *ghRepo) Create(ctx context.Context, genre *models.Genre) error {
	genre.ID = len(r.data) + 1
	genre.CreatedAt = time.Now()
	r.data[genre.ID] = genre
	return nil
}

func (r *ghRepo) Update(ctx context.Context, genre *models.Genre) error {
	if _, ok := r.data[genre.ID]; !ok {
		return sql.ErrNoRows
	}
	r.data[genre.ID] = genre
	return nil
}

func (r *ghRepo) Delete(ctx context.Context, id int) error {
	if _, ok := r.data[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.data, id)
	return nil
}

func TestGenreHandler_CRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newGHRepo()
	svc := service.NewGenreService(repo, validator.New())
	h := NewGenreHandler(svc)

	// seed
	existing := &models.Genre{ID: 1, Name: "Drama", CreatedAt: time.Now()}
	repo.data[existing.ID] = existing

	router := gin.New()
	router.GET("/genres", h.List)
	router.GET("/genres/:id", h.Get)
	router.POST("/genres", h.Create)
	router.PUT("/genres/:id", h.Update)
	router.DELETE("/genres/:id", h.Delete)

	// List
	{
		req := httptest.NewRequest(http.MethodGet, "/genres", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("list expected 200, got %d", w.Code)
		}
	}

	// Get not found
	{
		req := httptest.NewRequest(http.MethodGet, "/genres/9999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("get expected 404, got %d", w.Code)
		}
	}

	// Create ok
	{
		body, _ := json.Marshal(models.CreateGenreRequest{Name: "Comedy"})
		req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("create expected 201, got %d", w.Code)
		}
	}

	// Create duplicate
	{
		body, _ := json.Marshal(models.CreateGenreRequest{Name: "Drama"})
		req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("create duplicate expected 409, got %d", w.Code)
		}
	}

	// Update ok
	{
		body, _ := json.Marshal(models.CreateGenreRequest{Name: "Updated"})
		req := httptest.NewRequest(http.MethodPut, "/genres/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("update expected 200, got %d", w.Code)
		}
	}

	// Delete ok
	{
		req := httptest.NewRequest(http.MethodDelete, "/genres/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Fatalf("delete expected 204, got %d", w.Code)
		}
	}
}

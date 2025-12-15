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
	"github.com/google/uuid"

	"golang-project/internal/models"
	"golang-project/internal/service"
)

type mhMovieRepo struct {
	movies      map[uuid.UUID]*models.Movie
	movieGenres map[uuid.UUID][]uuid.UUID
}

type mhGenreLookup struct {
	data map[uuid.UUID]*models.Genre
}

func (r *mhGenreLookup) GetByID(ctx context.Context, id uuid.UUID) (*models.Genre, error) {
	if g, ok := r.data[id]; ok {
		return g, nil
	}
	return nil, sql.ErrNoRows
}

func newMHRepos() (*mhMovieRepo, *mhGenreLookup, uuid.UUID) {
	genreID := uuid.New()
	genreRepo := &mhGenreLookup{data: map[uuid.UUID]*models.Genre{
		genreID: {ID: genreID, Name: "Drama"},
	}}
	movieRepo := &mhMovieRepo{
		movies:      make(map[uuid.UUID]*models.Movie),
		movieGenres: make(map[uuid.UUID][]uuid.UUID),
	}
	return movieRepo, genreRepo, genreID
}

func (r *mhMovieRepo) List(ctx context.Context, filters models.MovieFilters, limit, offset int) ([]models.Movie, int, error) {
	result := make([]models.Movie, 0, len(r.movies))
	for _, m := range r.movies {
		result = append(result, *m)
	}
	return result, len(result), nil
}

func (r *mhMovieRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Movie, error) {
	if m, ok := r.movies[id]; ok {
		return m, nil
	}
	return nil, sql.ErrNoRows
}

func (r *mhMovieRepo) Create(ctx context.Context, movie *models.Movie) error {
	now := time.Now()
	movie.ID = uuid.New()
	movie.CreatedAt = now
	movie.UpdatedAt = now
	r.movies[movie.ID] = movie
	return nil
}

func (r *mhMovieRepo) Update(ctx context.Context, movie *models.Movie) error {
	if _, ok := r.movies[movie.ID]; !ok {
		return sql.ErrNoRows
	}
	movie.UpdatedAt = time.Now()
	r.movies[movie.ID] = movie
	return nil
}

func (r *mhMovieRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := r.movies[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.movies, id)
	delete(r.movieGenres, id)
	return nil
}

func (r *mhMovieRepo) SetGenres(ctx context.Context, movieID uuid.UUID, genreIDs []uuid.UUID) error {
	if _, ok := r.movies[movieID]; !ok {
		return sql.ErrNoRows
	}
	r.movieGenres[movieID] = genreIDs
	return nil
}

func (r *mhMovieRepo) GetGenresByMovieID(ctx context.Context, movieID uuid.UUID) ([]models.Genre, error) {
	ids := r.movieGenres[movieID]
	result := make([]models.Genre, 0, len(ids))
	for _, id := range ids {
		result = append(result, models.Genre{ID: id, Name: id.String()})
	}
	return result, nil
}

func TestMovieHandler_CRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mRepo, gRepo, genreID := newMHRepos()
	svc := service.NewMovieService(mRepo, gRepo, validator.New())
	h := NewMovieHandler(svc)

	router := gin.New()
	router.GET("/movies", h.List)
	router.GET("/movies/:id", h.Get)
	router.POST("/movies", h.Create)
	router.PUT("/movies/:id", h.Update)
	router.DELETE("/movies/:id", h.Delete)

	// Create ok
	createBody, _ := json.Marshal(models.CreateMovieRequest{
		Title:           "Movie",
		ReleaseYear:     2020,
		DurationMinutes: 100,
		GenreIDs:        []string{genreID.String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/movies", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create expected 201, got %d", w.Code)
	}

	// List
	req = httptest.NewRequest(http.MethodGet, "/movies", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d", w.Code)
	}

	// Get not found
	req = httptest.NewRequest(http.MethodGet, "/movies/"+uuid.New().String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get expected 404, got %d", w.Code)
	}
}

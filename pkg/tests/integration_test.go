package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"golang-project/internal/handler"
	"golang-project/internal/middleware"
	"golang-project/internal/models"
	"golang-project/internal/service"
	"golang-project/pkg/jwt"
)

// In-memory stubs for integration-style HTTP tests.

type memUserRepo struct {
	mu      sync.Mutex
	byID    map[uuid.UUID]*models.User
	byEmail map[string]*models.User
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{
		byID:    make(map[uuid.UUID]*models.User),
		byEmail: make(map[string]*models.User),
	}
}

func (r *memUserRepo) Create(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := uuid.New()
	now := time.Now()
	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	r.byID[id] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *memUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.byEmail[email]; ok {
		return u, nil
	}
	return nil, sql.ErrNoRows
}

type memGenreRepo struct {
	mu   sync.Mutex
	data map[uuid.UUID]*models.Genre
}

func newMemGenreRepo() *memGenreRepo {
	return &memGenreRepo{data: make(map[uuid.UUID]*models.Genre)}
}

func (r *memGenreRepo) GetAll(ctx context.Context) ([]models.Genre, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]models.Genre, 0, len(r.data))
	for _, g := range r.data {
		res = append(res, *g)
	}
	return res, nil
}

func (r *memGenreRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Genre, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.data[id]; ok {
		return g, nil
	}
	return nil, sql.ErrNoRows
}

func (r *memGenreRepo) GetByName(ctx context.Context, name string) (*models.Genre, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, g := range r.data {
		if g.Name == name {
			return g, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memGenreRepo) Create(ctx context.Context, genre *models.Genre) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := uuid.New()
	now := time.Now()
	genre.ID = id
	genre.CreatedAt = now
	r.data[id] = genre
	return nil
}

func (r *memGenreRepo) Update(ctx context.Context, genre *models.Genre) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[genre.ID]; !ok {
		return sql.ErrNoRows
	}
	r.data[genre.ID] = genre
	return nil
}

func (r *memGenreRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.data, id)
	return nil
}

type memMovieRepo struct {
	mu          sync.Mutex
	movies      map[uuid.UUID]*models.Movie
	movieGenres map[uuid.UUID][]uuid.UUID
}

func newMemMovieRepo() *memMovieRepo {
	return &memMovieRepo{
		movies:      make(map[uuid.UUID]*models.Movie),
		movieGenres: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (r *memMovieRepo) List(ctx context.Context, filters models.MovieFilters, limit, offset int) ([]models.Movie, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := make([]models.Movie, 0, len(r.movies))
	for _, m := range r.movies {
		all = append(all, *m)
	}
	total := len(all)
	if offset >= total {
		return []models.Movie{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (r *memMovieRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Movie, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.movies[id]; ok {
		return m, nil
	}
	return nil, sql.ErrNoRows
}

func (r *memMovieRepo) Create(ctx context.Context, movie *models.Movie) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := uuid.New()
	now := time.Now()
	movie.ID = id
	movie.CreatedAt = now
	movie.UpdatedAt = now
	r.movies[id] = movie
	return nil
}

func (r *memMovieRepo) Update(ctx context.Context, movie *models.Movie) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.movies[movie.ID]; !ok {
		return sql.ErrNoRows
	}
	movie.UpdatedAt = time.Now()
	r.movies[movie.ID] = movie
	return nil
}

func (r *memMovieRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.movies[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.movies, id)
	delete(r.movieGenres, id)
	return nil
}

func (r *memMovieRepo) SetGenres(ctx context.Context, movieID uuid.UUID, genreIDs []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.movies[movieID]; !ok {
		return sql.ErrNoRows
	}
	r.movieGenres[movieID] = genreIDs
	return nil
}

func (r *memMovieRepo) GetGenresByMovieID(ctx context.Context, movieID uuid.UUID) ([]models.Genre, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := r.movieGenres[movieID]
	res := make([]models.Genre, 0, len(ids))
	for _, id := range ids {
		res = append(res, models.Genre{ID: id, Name: id.String(), CreatedAt: time.Now()})
	}
	return res, nil
}

func (r *memMovieRepo) UpdateAverageRating(ctx context.Context, movieID uuid.UUID) error {
	return nil
}

type memReviewRepo struct {
	mu      sync.Mutex
	data    map[uuid.UUID]*models.Review
	byMovie map[uuid.UUID][]*models.Review
}

func newMemReviewRepo() *memReviewRepo {
	return &memReviewRepo{
		data:    make(map[uuid.UUID]*models.Review),
		byMovie: make(map[uuid.UUID][]*models.Review),
	}
}

func (r *memReviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rv, ok := r.data[id]; ok {
		return rv, nil
	}
	return nil, sql.ErrNoRows
}

func (r *memReviewRepo) GetByMovieAndUser(ctx context.Context, movieID, userID uuid.UUID) (*models.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rv := range r.byMovie[movieID] {
		if rv.UserID == userID {
			return rv, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memReviewRepo) GetByMovieID(ctx context.Context, movieID uuid.UUID, limit, offset int) ([]models.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := r.byMovie[movieID]
	total := len(all)
	if offset >= total {
		return []models.Review{}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	res := make([]models.Review, 0, end-offset)
	for _, rv := range all[offset:end] {
		res = append(res, *rv)
	}
	return res, nil
}

func (r *memReviewRepo) Create(ctx context.Context, review *models.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := uuid.New()
	now := time.Now()
	review.ID = id
	review.CreatedAt = now
	review.UpdatedAt = now
	r.data[id] = review
	r.byMovie[review.MovieID] = append(r.byMovie[review.MovieID], review)
	return nil
}

func (r *memReviewRepo) Update(ctx context.Context, review *models.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[review.ID]; !ok {
		return sql.ErrNoRows
	}
	review.UpdatedAt = time.Now()
	r.data[review.ID] = review
	for i, rv := range r.byMovie[review.MovieID] {
		if rv.ID == review.ID {
			r.byMovie[review.MovieID][i] = review
			break
		}
	}
	return nil
}

func (r *memReviewRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	review, ok := r.data[id]
	if !ok {
		return sql.ErrNoRows
	}
	delete(r.data, id)
	arr := r.byMovie[review.MovieID]
	for i, rv := range arr {
		if rv.ID == id {
			arr = append(arr[:i], arr[i+1:]...)
			break
		}
	}
	r.byMovie[review.MovieID] = arr
	return nil
}

func buildTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	validator := validator.New()
	secret := "test-secret"

	userRepo := newMemUserRepo()
	genreRepo := newMemGenreRepo()
	movieRepo := newMemMovieRepo()
	reviewRepo := newMemReviewRepo()

	authSvc := service.NewAuthService(userRepo, validator, secret)
	genreSvc := service.NewGenreService(genreRepo, validator)
	movieSvc := service.NewMovieService(movieRepo, genreRepo, validator)
	reviewSvc := service.NewReviewService(reviewRepo, movieRepo, validator)

	authH := handler.NewAuthHandler(authSvc)
	genreH := handler.NewGenreHandler(genreSvc)
	movieH := handler.NewMovieHandler(movieSvc)
	reviewH := handler.NewReviewHandler(reviewSvc)

	router := gin.New()
	router.Use(middleware.Logger(), gin.Recovery(), middleware.RequestID(), middleware.CORS(), middleware.BodyLimit(1<<20))

	api := router.Group("/api/v1")

	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)
	api.GET("/genres", genreH.List)
	api.GET("/genres/:id", genreH.Get)
	api.GET("/movies", movieH.List)
	api.GET("/movies/:id", movieH.Get)
	api.GET("/movies/:id/reviews", reviewH.ListByMovie)

	admin := api.Group("/", middleware.AuthMiddleware(secret), middleware.RequireRoles("admin"))
	admin.POST("/genres", genreH.Create)
	admin.POST("/movies", movieH.Create)

	protected := api.Group("/", middleware.AuthMiddleware(secret))
	protected.POST("/movies/:id/reviews", reviewH.Create)
	protected.PUT("/reviews/:id", reviewH.Update)
	protected.DELETE("/reviews/:id", reviewH.Delete)

	// seed admin user
	hash, err := jwt.HashPassword("adminpass")
	if err != nil {
		t.Fatalf("hash admin password: %v", err)
	}
	_ = userRepo.Create(context.Background(), &models.User{
		Email:        "admin@example.com",
		Username:     "admin",
		PasswordHash: hash,
		Role:         "admin",
	})

	return router
}

func TestIntegration_Flow(t *testing.T) {
	router := buildTestRouter(t)

	// Login as admin
	adminToken := login(t, router, "admin@example.com", "adminpass")

	// Create genre
	genreID := createGenre(t, router, adminToken, "Drama")

	// Create movie
	movieID := createMovie(t, router, adminToken, "Inception", genreID)

	// Register user and login
	register(t, router, "user@example.com", "user", "password123")
	userToken := login(t, router, "user@example.com", "password123")

	// Create review
	reviewID := createReview(t, router, userToken, movieID, 9, "Great", "Nice")

	// Update review
	updateReview(t, router, userToken, reviewID, 8, "Updated", "Still good")

	// Delete review as admin
	deleteReview(t, router, adminToken, reviewID)
}

func login(t *testing.T, r *gin.Engine, email, password string) string {
	body, _ := json.Marshal(models.LoginRequest{Email: email, Password: password})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login expected 200, got %d body %s", w.Code, w.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Token == "" {
		t.Fatalf("login parse token err=%v body=%s", err, w.Body.String())
	}
	return resp.Token
}

func register(t *testing.T, r *gin.Engine, email, username, password string) {
	body, _ := json.Marshal(models.CreateUserRequest{Email: email, Username: username, Password: password})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register expected 201, got %d body %s", w.Code, w.Body.String())
	}
}

func createGenre(t *testing.T, r *gin.Engine, token, name string) string {
	body, _ := json.Marshal(models.CreateGenreRequest{Name: name})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/genres", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create genre expected 201, got %d body %s", w.Code, w.Body.String())
	}
	var resp models.Genre
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.ID == uuid.Nil {
		t.Fatalf("parse genre err=%v body=%s", err, w.Body.String())
	}
	return resp.ID.String()
}

func createMovie(t *testing.T, r *gin.Engine, token, title, genreID string) string {
	body, _ := json.Marshal(models.CreateMovieRequest{
		Title:           title,
		ReleaseYear:     2020,
		DurationMinutes: 100,
		GenreIDs:        []string{genreID},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/movies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create movie expected 201, got %d body %s", w.Code, w.Body.String())
	}
	var resp models.Movie
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.ID == uuid.Nil {
		t.Fatalf("parse movie err=%v body=%s", err, w.Body.String())
	}
	return resp.ID.String()
}

func createReview(t *testing.T, r *gin.Engine, token, movieID string, rating int, title, content string) string {
	body, _ := json.Marshal(models.CreateReviewRequest{Rating: rating, Title: title, Content: content})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/movies/"+movieID+"/reviews", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create review expected 201, got %d body %s", w.Code, w.Body.String())
	}
	var resp models.Review
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.ID == uuid.Nil {
		t.Fatalf("parse review err=%v body=%s", err, w.Body.String())
	}
	return resp.ID.String()
}

func updateReview(t *testing.T, r *gin.Engine, token, reviewID string, rating int, title, content string) {
	body, _ := json.Marshal(models.UpdateReviewRequest{Rating: rating, Title: title, Content: content})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/reviews/"+reviewID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update review expected 200, got %d body %s", w.Code, w.Body.String())
	}
}

func deleteReview(t *testing.T, r *gin.Engine, token, reviewID string) {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+reviewID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete review expected 204, got %d body %s", w.Code, w.Body.String())
	}
}

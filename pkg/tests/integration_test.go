package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"golang-project/internal/handler"
	"golang-project/internal/middleware"
	"golang-project/internal/models"
	"golang-project/internal/service"
	"golang-project/pkg/jwt"
)

// In-memory stubs for integration-style HTTP tests.

type memUserRepo struct {
	mu      sync.Mutex
	byID    map[int]*models.User
	byEmail map[string]*models.User
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{
		byID:    make(map[int]*models.User),
		byEmail: make(map[string]*models.User),
	}
}

func (r *memUserRepo) Create(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := len(r.byID) + 1
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

func (r *memUserRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.byID[id]; ok {
		return u, nil
	}
	return nil, sql.ErrNoRows
}

func (r *memUserRepo) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]models.User, 0, len(r.byID))
	for _, u := range r.byID {
		result = append(result, *u)
	}
	total := len(result)
	if offset >= total {
		return []models.User{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return result[offset:end], total, nil
}

func (r *memUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.byID {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memUserRepo) UpdateRole(ctx context.Context, id int, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.byID[id]; ok {
		u.Role = role
		return nil
	}
	return sql.ErrNoRows
}

func (r *memUserRepo) Update(ctx context.Context, id int, email, username string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return sql.ErrNoRows
	}
	if email != "" && email != u.Email {
		delete(r.byEmail, u.Email)
		u.Email = email
		r.byEmail[email] = u
	}
	if username != "" {
		u.Username = username
	}
	u.UpdatedAt = time.Now()
	return nil
}

func (r *memUserRepo) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return sql.ErrNoRows
	}
	delete(r.byID, id)
	delete(r.byEmail, u.Email)
	return nil
}

type memGenreRepo struct {
	mu   sync.Mutex
	data map[int]*models.Genre
}

func newMemGenreRepo() *memGenreRepo {
	return &memGenreRepo{data: make(map[int]*models.Genre)}
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

func (r *memGenreRepo) GetByID(ctx context.Context, id int) (*models.Genre, error) {
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
	id := len(r.data) + 1
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

func (r *memGenreRepo) Delete(ctx context.Context, id int) error {
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
	movies      map[int]*models.Movie
	movieGenres map[int][]int
}

func newMemMovieRepo() *memMovieRepo {
	return &memMovieRepo{
		movies:      make(map[int]*models.Movie),
		movieGenres: make(map[int][]int),
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

func (r *memMovieRepo) GetByID(ctx context.Context, id int) (*models.Movie, error) {
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
	id := len(r.movies) + 1
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

func (r *memMovieRepo) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.movies[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.movies, id)
	delete(r.movieGenres, id)
	return nil
}

func (r *memMovieRepo) SetGenres(ctx context.Context, movieID int, genreIDs []int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.movies[movieID]; !ok {
		return sql.ErrNoRows
	}
	r.movieGenres[movieID] = genreIDs
	return nil
}

func (r *memMovieRepo) GetGenresByMovieID(ctx context.Context, movieID int) ([]models.Genre, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := r.movieGenres[movieID]
	res := make([]models.Genre, 0, len(ids))
	for _, id := range ids {
		res = append(res, models.Genre{ID: id, Name: "", CreatedAt: time.Now()})
	}
	return res, nil
}

func (r *memMovieRepo) UpdateAverageRating(ctx context.Context, movieID int) error {
	return nil
}

type memReviewRepo struct {
	mu      sync.Mutex
	data    map[int]*models.Review
	byMovie map[int][]*models.Review
}

func newMemReviewRepo() *memReviewRepo {
	return &memReviewRepo{
		data:    make(map[int]*models.Review),
		byMovie: make(map[int][]*models.Review),
	}
}

func (r *memReviewRepo) GetByID(ctx context.Context, id int) (*models.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rv, ok := r.data[id]; ok {
		return rv, nil
	}
	return nil, sql.ErrNoRows
}

func (r *memReviewRepo) GetByMovieAndUser(ctx context.Context, movieID, userID int) (*models.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rv := range r.byMovie[movieID] {
		if rv.UserID == userID {
			return rv, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memReviewRepo) GetByMovieID(ctx context.Context, movieID int, filters models.ReviewFilters, limit, offset int) ([]models.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := r.byMovie[movieID]
	filtered := make([]*models.Review, 0, len(all))
	for _, rv := range all {
		if filters.MinRating > 0 && rv.Rating < filters.MinRating {
			continue
		}
		if filters.MaxRating > 0 && rv.Rating > filters.MaxRating {
			continue
		}
		filtered = append(filtered, rv)
	}
	all = filtered
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

func (r *memReviewRepo) GetByUserID(ctx context.Context, userID int, filters models.ReviewFilters, limit, offset int) ([]models.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := make([]*models.Review, 0, len(r.data))
	for _, rv := range r.data {
		if rv.UserID == userID {
			if filters.MinRating > 0 && rv.Rating < filters.MinRating {
				continue
			}
			if filters.MaxRating > 0 && rv.Rating > filters.MaxRating {
				continue
			}
			all = append(all, rv)
		}
	}
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
	id := len(r.data) + 1
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

func (r *memReviewRepo) Delete(ctx context.Context, id int) error {
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

func (r *memReviewRepo) CountByUserID(ctx context.Context, userID int) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, rv := range r.data {
		if rv.UserID == userID {
			count++
		}
	}
	return count, nil
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
	reviewSvc := service.NewReviewService(reviewRepo, movieRepo, validator, nil)

	authH := handler.NewAuthHandler(authSvc)
	genreH := handler.NewGenreHandler(genreSvc)
	movieH := handler.NewMovieHandler(movieSvc)
	reviewH := handler.NewReviewHandler(reviewSvc)
	userH := handler.NewUserHandler(service.NewUserService(userRepo, validator), reviewSvc)

	router := gin.New()
	router.Use(middleware.Logger(), gin.Recovery(), middleware.RequestID(), middleware.CORS(), middleware.BodyLimit(1<<20))

	api := router.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)
	api.GET("/genres", genreH.List)
	api.GET("/genres/:id", genreH.Get)
	api.GET("/movies", movieH.List)
	api.GET("/movies/:id", movieH.Get)
	api.GET("/movies/:id/reviews", reviewH.ListByMovie)
	api.GET("/users/:id/reviews", userH.UserReviews)

	admin := api.Group("/", middleware.AuthMiddleware(secret), middleware.RequireRoles("admin"))
	admin.GET("/users", userH.ListUsers)
	admin.GET("/users/:id", userH.GetUser)
	admin.PUT("/users/:id", userH.UpdateUser)
	admin.PUT("/users/:id/role", userH.UpdateRole)
	admin.DELETE("/users/:id", userH.DeleteUser)
	admin.POST("/genres", genreH.Create)
	admin.PUT("/genres/:id", genreH.Update)
	admin.DELETE("/genres/:id", genreH.Delete)
	admin.POST("/movies", movieH.Create)
	admin.PUT("/movies/:id", movieH.Update)
	admin.DELETE("/movies/:id", movieH.Delete)

	protected := api.Group("/", middleware.AuthMiddleware(secret))
	protected.GET("/me", userH.Me)
	protected.GET("/me/reviews", userH.MyReviews)
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

func TestIntegration_ReviewCreate_Unauthorized(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	genreID := createGenre(t, router, adminToken, "Drama")
	movieID := createMovie(t, router, adminToken, "Inception", genreID)

	body, _ := json.Marshal(models.CreateReviewRequest{Rating: 9, Title: "Title", Content: "Body"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/movies/"+movieID+"/reviews", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("create review without token expected 401, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_ReviewUpdate_ForbiddenForOtherUser(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	genreID := createGenre(t, router, adminToken, "Drama")
	movieID := createMovie(t, router, adminToken, "Inception", genreID)

	register(t, router, "user1@example.com", "user1", "password123")
	user1Token := login(t, router, "user1@example.com", "password123")

	register(t, router, "user2@example.com", "user2", "password123")
	user2Token := login(t, router, "user2@example.com", "password123")

	reviewID := createReview(t, router, user1Token, movieID, 9, "Title", "Body")

	body, _ := json.Marshal(models.UpdateReviewRequest{Rating: 5, Title: "Other user", Content: "Should be forbidden"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/reviews/"+reviewID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("update review by other user expected 403, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AdminEndpoints_ForbiddenForUser(t *testing.T) {
	router := buildTestRouter(t)

	register(t, router, "user@example.com", "user", "password123")
	userToken := login(t, router, "user@example.com", "password123")

	body, _ := json.Marshal(models.CreateGenreRequest{Name: "UserGenre"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/genres", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("create genre as user expected 403, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_Health(t *testing.T) {
	router := buildTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("health expected 200, got %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse health response err=%v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("health status expected 'ok', got %v", resp["status"])
	}
}

func TestIntegration_Me(t *testing.T) {
	router := buildTestRouter(t)

	register(t, router, "user@example.com", "user", "password123")
	userToken := login(t, router, "user@example.com", "password123")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("me expected 200, got %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse me response err=%v", err)
	}
	if resp["user"] == nil {
		t.Fatalf("me response should contain 'user'")
	}
}

func TestIntegration_MeReviews(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	genreID := createGenre(t, router, adminToken, "Drama")
	movieID := createMovie(t, router, adminToken, "Inception", genreID)

	register(t, router, "user@example.com", "user", "password123")
	userToken := login(t, router, "user@example.com", "password123")

	createReview(t, router, userToken, movieID, 9, "Great", "Nice")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/reviews", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("me/reviews expected 200, got %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse me/reviews response err=%v", err)
	}
	if resp["data"] == nil {
		t.Fatalf("me/reviews response should contain 'data'")
	}
}

func TestIntegration_UserReviews(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	genreID := createGenre(t, router, adminToken, "Drama")
	movieID := createMovie(t, router, adminToken, "Inception", genreID)

	register(t, router, "user@example.com", "user", "password123")
	userToken := login(t, router, "user@example.com", "password123")

	createReview(t, router, userToken, movieID, 9, "Great", "Nice")

	var userResp map[string]interface{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &userResp)
	userData := userResp["user"].(map[string]interface{})
	userID := strconv.Itoa(int(userData["id"].(float64)))

	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID+"/reviews", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("users/:id/reviews expected 200, got %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse users/:id/reviews response err=%v", err)
	}
	if resp["data"] == nil {
		t.Fatalf("users/:id/reviews response should contain 'data'")
	}
}

func TestIntegration_AdminUsers(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	register(t, router, "user@example.com", "user", "password123")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("admin GET /users expected 200, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AdminGetUser(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	register(t, router, "user@example.com", "user", "password123")

	var userResp map[string]interface{}
	userToken := login(t, router, "user@example.com", "password123")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &userResp)
	userData := userResp["user"].(map[string]interface{})
	userID := strconv.Itoa(int(userData["id"].(float64)))

	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("admin GET /users/:id expected 200, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AdminUpdateRole(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	register(t, router, "user@example.com", "user", "password123")

	var userResp map[string]interface{}
	userToken := login(t, router, "user@example.com", "password123")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &userResp)
	userData := userResp["user"].(map[string]interface{})
	userID := strconv.Itoa(int(userData["id"].(float64)))

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/users/"+userID+"/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("admin PUT /users/:id/role expected 204, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AdminGenresUpdateDelete(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	genreID := createGenre(t, router, adminToken, "Drama")

	body, _ := json.Marshal(models.CreateGenreRequest{Name: "Updated Drama"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/genres/"+genreID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("admin PUT /genres/:id expected 200, got %d body %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/genres/"+genreID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE /genres/:id expected 204, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AdminMoviesUpdateDelete(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	genreID := createGenre(t, router, adminToken, "Drama")
	movieID := createMovie(t, router, adminToken, "Inception", genreID)

	body, _ := json.Marshal(models.CreateMovieRequest{
		Title:           "Updated Inception",
		ReleaseYear:     2021,
		DurationMinutes: 120,
		GenreIDs:        []string{genreID},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/movies/"+movieID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("admin PUT /movies/:id expected 200, got %d body %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/movies/"+movieID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE /movies/:id expected 204, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_MovieReviewsList(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	genreID := createGenre(t, router, adminToken, "Drama")
	movieID := createMovie(t, router, adminToken, "Inception", genreID)

	register(t, router, "user@example.com", "user", "password123")
	userToken := login(t, router, "user@example.com", "password123")

	createReview(t, router, userToken, movieID, 9, "Great", "Nice")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/"+movieID+"/reviews", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /movies/:id/reviews expected 200, got %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse /movies/:id/reviews response err=%v", err)
	}
	if resp["data"] == nil {
		t.Fatalf("/movies/:id/reviews response should contain 'data'")
	}
}

func TestIntegration_AdminUpdateUser(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	register(t, router, "user@example.com", "user", "password123")

	var userResp map[string]interface{}
	userToken := login(t, router, "user@example.com", "password123")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &userResp)
	userData := userResp["user"].(map[string]interface{})
	userID := strconv.Itoa(int(userData["id"].(float64)))

	body, _ := json.Marshal(models.UpdateUserRequest{
		Email:    "updated@example.com",
		Username: "updateduser",
	})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/users/"+userID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("admin PUT /users/:id expected 200, got %d body %s", w.Code, w.Body.String())
	}
	var updatedUser models.User
	if err := json.Unmarshal(w.Body.Bytes(), &updatedUser); err != nil {
		t.Fatalf("parse updated user err=%v", err)
	}
	if updatedUser.Email != "updated@example.com" {
		t.Fatalf("expected email updated@example.com, got %s", updatedUser.Email)
	}
	if updatedUser.Username != "updateduser" {
		t.Fatalf("expected username updateduser, got %s", updatedUser.Username)
	}
}

func TestIntegration_AdminDeleteUser(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")
	register(t, router, "user@example.com", "user", "password123")

	var userResp map[string]interface{}
	userToken := login(t, router, "user@example.com", "password123")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &userResp)
	userData := userResp["user"].(map[string]interface{})
	userID := strconv.Itoa(int(userData["id"].(float64)))

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE /users/:id expected 204, got %d body %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("get deleted user expected 404, got %d body %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AdminDeleteSelf_Forbidden(t *testing.T) {
	router := buildTestRouter(t)

	adminToken := login(t, router, "admin@example.com", "adminpass")

	var adminResp map[string]interface{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &adminResp)
	adminData := adminResp["user"].(map[string]interface{})
	adminID := strconv.Itoa(int(adminData["id"].(float64)))

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+adminID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("admin DELETE self expected 400, got %d body %s", w.Code, w.Body.String())
	}
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
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.ID == 0 {
		t.Fatalf("parse genre err=%v body=%s", err, w.Body.String())
	}
	return strconv.Itoa(resp.ID)
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
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.ID == 0 {
		t.Fatalf("parse movie err=%v body=%s", err, w.Body.String())
	}
	return strconv.Itoa(resp.ID)
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
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.ID == 0 {
		t.Fatalf("parse review err=%v body=%s", err, w.Body.String())
	}
	return strconv.Itoa(resp.ID)
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

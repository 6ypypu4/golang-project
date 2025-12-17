package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"golang-project/internal/models"
)

// MockMovieRepository implements MovieRepository for testing
type MockMovieRepository struct {
	movies   map[int]*models.Movie
	genres   map[int][]int // movie_id -> []genre_id
	nextID   int
	nextUser int
}

func NewMockMovieRepository() *MockMovieRepository {
	return &MockMovieRepository{
		movies: make(map[int]*models.Movie),
		genres: make(map[int][]int),
		nextID: 1,
	}
}

func (r *MockMovieRepository) GetByID(ctx context.Context, id int) (*models.Movie, error) {
	if movie, exists := r.movies[id]; exists {
		// Add genres to the movie
		if genreIDs, exists := r.genres[id]; exists {
			movie.Genres = make([]models.Genre, len(genreIDs))
			for i, genreID := range genreIDs {
				movie.Genres[i] = models.Genre{ID: genreID, Name: fmt.Sprintf("Genre%d", genreID)}
			}
		}
		return movie, nil
	}
	return nil, sql.ErrNoRows
}

func (r *MockMovieRepository) Create(ctx context.Context, movie *models.Movie) error {
	movie.ID = r.nextID
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = movie.CreatedAt
	r.movies[movie.ID] = movie
	r.nextID++
	return nil
}

func (r *MockMovieRepository) Update(ctx context.Context, movie *models.Movie) error {
	if _, exists := r.movies[movie.ID]; !exists {
		return sql.ErrNoRows
	}
	movie.UpdatedAt = time.Now()
	r.movies[movie.ID] = movie
	return nil
}

func (r *MockMovieRepository) Delete(ctx context.Context, id int) error {
	if _, exists := r.movies[id]; !exists {
		return sql.ErrNoRows
	}
	delete(r.movies, id)
	delete(r.genres, id)
	return nil
}

func (r *MockMovieRepository) List(ctx context.Context, filters models.MovieFilters, limit, offset int) ([]models.Movie, int, error) {
	result := make([]models.Movie, 0)
	for _, movie := range r.movies {
		if r.matchesFilters(movie, filters) {
			result = append(result, *movie)
		}
	}

	// Apply pagination
	total := len(result)
	start := offset
	if start >= total {
		return []models.Movie{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}

	paginated := result[start:end]
	for i := range paginated {
		// Add genres to each movie
		if genreIDs, exists := r.genres[paginated[i].ID]; exists {
			paginated[i].Genres = make([]models.Genre, len(genreIDs))
			for j, genreID := range genreIDs {
				paginated[i].Genres[j] = models.Genre{ID: genreID, Name: fmt.Sprintf("Genre%d", genreID)}
			}
		}
	}

	return paginated, total, nil
}

func (r *MockMovieRepository) GetGenresByMovieID(ctx context.Context, movieID int) ([]models.Genre, error) {
	if genreIDs, exists := r.genres[movieID]; exists {
		genres := make([]models.Genre, len(genreIDs))
		for i, genreID := range genreIDs {
			genres[i] = models.Genre{ID: genreID, Name: fmt.Sprintf("Genre%d", genreID)}
		}
		return genres, nil
	}
	return []models.Genre{}, nil
}

func (r *MockMovieRepository) SetGenres(ctx context.Context, movieID int, genreIDs []int) error {
	if _, exists := r.movies[movieID]; !exists {
		return sql.ErrNoRows
	}
	r.genres[movieID] = genreIDs
	return nil
}

func (r *MockMovieRepository) UpdateAverageRating(ctx context.Context, movieID int) error {
	// Mock implementation - in real implementation this would calculate from reviews
	return nil
}

func (r *MockMovieRepository) Count(ctx context.Context) (int, error) {
	return len(r.movies), nil
}

func (r *MockMovieRepository) matchesFilters(movie *models.Movie, filters models.MovieFilters) bool {
	// Check genre filter
	if filters.GenreID != nil {
		genreIDs, exists := r.genres[movie.ID]
		if !exists {
			return false
		}
		found := false
		for _, gid := range genreIDs {
			if gid == *filters.GenreID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check year filter
	if filters.Year != 0 && movie.ReleaseYear != filters.Year {
		return false
	}

	// Check minimum rating filter
	if filters.MinRating > 0 && movie.AverageRating < filters.MinRating {
		return false
	}

	// Check search filter
	if filters.Search != "" {
		search := filters.Search
		titleMatch := containsIgnoreCase(movie.Title, search)
		descriptionMatch := containsIgnoreCase(movie.Description, search)
		if !titleMatch && !descriptionMatch {
			return false
		}
	}

	return true
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && findSubstringIgnoreCase(s, substr)
}

func findSubstringIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+'a'-'A')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func TestMovieRepository_GetByID(t *testing.T) {
	repo := NewMockMovieRepository()
	ctx := context.Background()

	// Test getting non-existent movie
	_, err := repo.GetByID(ctx, 1)
	if err == nil {
		t.Error("Expected error for non-existent movie")
	}

	// Create a movie
	movie := &models.Movie{
		Title:           "Test Movie",
		Description:     "Test Description",
		ReleaseYear:     2023,
		Director:        "Test Director",
		DurationMinutes: 120,
	}
	err = repo.Create(ctx, movie)
	if err != nil {
		t.Errorf("Unexpected error creating movie: %v", err)
	}

	// Test getting existing movie
	retrieved, err := repo.GetByID(ctx, movie.ID)
	if err != nil {
		t.Errorf("Unexpected error getting movie: %v", err)
	}
	if retrieved.ID != movie.ID || retrieved.Title != movie.Title {
		t.Errorf("Retrieved movie doesn't match. Expected: %+v, Got: %+v", movie, retrieved)
	}
}

func TestMovieRepository_Create(t *testing.T) {
	repo := NewMockMovieRepository()
	ctx := context.Background()

	movie := &models.Movie{
		Title:           "New Movie",
		Description:     "New Description",
		ReleaseYear:     2024,
		Director:        "New Director",
		DurationMinutes: 150,
	}
	err := repo.Create(ctx, movie)
	if err != nil {
		t.Errorf("Unexpected error creating movie: %v", err)
	}

	if movie.ID == 0 {
		t.Error("Expected movie ID to be set")
	}
	if movie.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if movie.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, movie.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving created movie: %v", err)
	}
	if retrieved.Title != movie.Title {
		t.Errorf("Movie title mismatch. Expected: %s, Got: %s", movie.Title, retrieved.Title)
	}
}

func TestMovieRepository_Update(t *testing.T) {
	repo := NewMockMovieRepository()
	ctx := context.Background()

	// Create a movie
	movie := &models.Movie{
		Title:           "Original Title",
		Description:     "Original Description",
		ReleaseYear:     2023,
		Director:        "Original Director",
		DurationMinutes: 120,
	}
	err := repo.Create(ctx, movie)
	if err != nil {
		t.Errorf("Unexpected error creating movie: %v", err)
	}

	// Update the movie
	originalUpdatedAt := movie.UpdatedAt
	movie.Title = "Updated Title"
	movie.Description = "Updated Description"
	err = repo.Update(ctx, movie)
	if err != nil {
		t.Errorf("Unexpected error updating movie: %v", err)
	}

	// Verify the update
	retrieved, err := repo.GetByID(ctx, movie.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving updated movie: %v", err)
	}
	if retrieved.Title != "Updated Title" {
		t.Errorf("Movie title not updated. Expected: Updated Title, Got: %s", retrieved.Title)
	}
	if retrieved.UpdatedAt.Before(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}

	// Test updating non-existent movie
	nonExistent := &models.Movie{ID: 999, Title: "Test"}
	err = repo.Update(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error updating non-existent movie")
	}
}

func TestMovieRepository_Delete(t *testing.T) {
	repo := NewMockMovieRepository()
	ctx := context.Background()

	// Create a movie
	movie := &models.Movie{
		Title:           "To Delete",
		Description:     "Description",
		ReleaseYear:     2023,
		Director:        "Director",
		DurationMinutes: 120,
	}
	err := repo.Create(ctx, movie)
	if err != nil {
		t.Errorf("Unexpected error creating movie: %v", err)
	}

	// Delete the movie
	err = repo.Delete(ctx, movie.ID)
	if err != nil {
		t.Errorf("Unexpected error deleting movie: %v", err)
	}

	// Verify it was deleted
	_, err = repo.GetByID(ctx, movie.ID)
	if err == nil {
		t.Error("Expected error retrieving deleted movie")
	}

	// Test deleting non-existent movie
	err = repo.Delete(ctx, 999)
	if err == nil {
		t.Error("Expected error deleting non-existent movie")
	}
}

func TestMovieRepository_List(t *testing.T) {
	repo := NewMockMovieRepository()
	ctx := context.Background()

	// Test empty repository
	movies, total, err := repo.List(ctx, models.MovieFilters{}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error listing movies: %v", err)
	}
	if len(movies) != 0 || total != 0 {
		t.Errorf("Expected empty list, got %d movies with total %d", len(movies), total)
	}

	// Create some movies
	movie1 := &models.Movie{
		Title:           "Action Movie",
		Description:     "An action-packed movie",
		ReleaseYear:     2023,
		Director:        "Action Director",
		DurationMinutes: 120,
		AverageRating:   8.5,
	}
	movie2 := &models.Movie{
		Title:           "Comedy Movie",
		Description:     "A funny comedy",
		ReleaseYear:     2022,
		Director:        "Comedy Director",
		DurationMinutes: 90,
		AverageRating:   7.2,
	}
	err = repo.Create(ctx, movie1)
	if err != nil {
		t.Errorf("Unexpected error creating movie1: %v", err)
	}
	err = repo.Create(ctx, movie2)
	if err != nil {
		t.Errorf("Unexpected error creating movie2: %v", err)
	}

	// Test listing all movies
	movies, total, err = repo.List(ctx, models.MovieFilters{}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error listing movies: %v", err)
	}
	if len(movies) != 2 || total != 2 {
		t.Errorf("Expected 2 movies, got %d with total %d", len(movies), total)
	}

	// Test year filter
	movies, total, err = repo.List(ctx, models.MovieFilters{Year: 2023}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error listing movies: %v", err)
	}
	if len(movies) != 1 || total != 1 {
		t.Errorf("Expected 1 movie from 2023, got %d with total %d", len(movies), total)
	}
	if movies[0].Title != "Action Movie" {
		t.Errorf("Expected Action Movie, got %s", movies[0].Title)
	}

	// Test minimum rating filter
	movies, total, err = repo.List(ctx, models.MovieFilters{MinRating: 8.0}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error listing movies: %v", err)
	}
	if len(movies) != 1 || total != 1 {
		t.Errorf("Expected 1 movie with rating >= 8.0, got %d with total %d", len(movies), total)
	}
	if movies[0].Title != "Action Movie" {
		t.Errorf("Expected Action Movie, got %s", movies[0].Title)
	}

	// Test search filter
	movies, total, err = repo.List(ctx, models.MovieFilters{Search: "action"}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error listing movies: %v", err)
	}
	if len(movies) != 1 || total != 1 {
		t.Errorf("Expected 1 movie matching 'action', got %d with total %d", len(movies), total)
	}
	if movies[0].Title != "Action Movie" {
		t.Errorf("Expected Action Movie, got %s", movies[0].Title)
	}
}

func TestMovieRepository_SetGenres(t *testing.T) {
	repo := NewMockMovieRepository()
	ctx := context.Background()

	// Create a movie
	movie := &models.Movie{
		Title:           "Test Movie",
		Description:     "Test Description",
		ReleaseYear:     2023,
		Director:        "Test Director",
		DurationMinutes: 120,
	}
	err := repo.Create(ctx, movie)
	if err != nil {
		t.Errorf("Unexpected error creating movie: %v", err)
	}

	// Set genres
	genreIDs := []int{1, 2, 3}
	err = repo.SetGenres(ctx, movie.ID, genreIDs)
	if err != nil {
		t.Errorf("Unexpected error setting genres: %v", err)
	}

	// Verify genres were set
	retrieved, err := repo.GetByID(ctx, movie.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving movie: %v", err)
	}
	if len(retrieved.Genres) != 3 {
		t.Errorf("Expected 3 genres, got %d", len(retrieved.Genres))
	}

	// Test setting genres for non-existent movie
	err = repo.SetGenres(ctx, 999, []int{1, 2})
	if err == nil {
		t.Error("Expected error setting genres for non-existent movie")
	}
}

func TestMovieRepository_Count(t *testing.T) {
	repo := NewMockMovieRepository()
	ctx := context.Background()

	// Test empty repository
	count, err := repo.Count(ctx)
	if err != nil {
		t.Errorf("Unexpected error counting movies: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Create some movies
	movie1 := &models.Movie{
		Title:           "Movie 1",
		Description:     "Description 1",
		ReleaseYear:     2023,
		Director:        "Director 1",
		DurationMinutes: 120,
	}
	movie2 := &models.Movie{
		Title:           "Movie 2",
		Description:     "Description 2",
		ReleaseYear:     2024,
		Director:        "Director 2",
		DurationMinutes: 150,
	}
	err = repo.Create(ctx, movie1)
	if err != nil {
		t.Errorf("Unexpected error creating movie1: %v", err)
	}
	err = repo.Create(ctx, movie2)
	if err != nil {
		t.Errorf("Unexpected error creating movie2: %v", err)
	}

	// Test counting movies
	count, err = repo.Count(ctx)
	if err != nil {
		t.Errorf("Unexpected error counting movies: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

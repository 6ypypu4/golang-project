package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"golang-project/internal/models"
)

// MockGenreRepository implements GenreRepository for testing
type MockGenreRepository struct {
	genres map[int]*models.Genre
	nextID int
}

func NewMockGenreRepository() *MockGenreRepository {
	return &MockGenreRepository{
		genres: make(map[int]*models.Genre),
		nextID: 1,
	}
}

func (r *MockGenreRepository) GetByID(ctx context.Context, id int) (*models.Genre, error) {
	if genre, exists := r.genres[id]; exists {
		return genre, nil
	}
	return nil, sql.ErrNoRows
}

func (r *MockGenreRepository) GetByName(ctx context.Context, name string) (*models.Genre, error) {
	for _, genre := range r.genres {
		if genre.Name == name {
			return genre, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *MockGenreRepository) Create(ctx context.Context, genre *models.Genre) error {
	genre.ID = r.nextID
	genre.CreatedAt = time.Now()
	r.genres[genre.ID] = genre
	r.nextID++
	return nil
}

func (r *MockGenreRepository) Update(ctx context.Context, genre *models.Genre) error {
	if _, exists := r.genres[genre.ID]; !exists {
		return sql.ErrNoRows
	}
	r.genres[genre.ID] = genre
	return nil
}

func (r *MockGenreRepository) Delete(ctx context.Context, id int) error {
	if _, exists := r.genres[id]; !exists {
		return sql.ErrNoRows
	}
	delete(r.genres, id)
	return nil
}

func (r *MockGenreRepository) GetAll(ctx context.Context) ([]models.Genre, error) {
	result := make([]models.Genre, 0, len(r.genres))
	for _, genre := range r.genres {
		result = append(result, *genre)
	}
	return result, nil
}

func (r *MockGenreRepository) Count(ctx context.Context) (int, error) {
	return len(r.genres), nil
}

func TestGenreRepository_GetByID(t *testing.T) {
	repo := NewMockGenreRepository()
	ctx := context.Background()

	// Test getting non-existent genre
	_, err := repo.GetByID(ctx, 1)
	if err == nil {
		t.Error("Expected error for non-existent genre")
	}

	// Create a genre
	genre := &models.Genre{Name: "Action"}
	err = repo.Create(ctx, genre)
	if err != nil {
		t.Errorf("Unexpected error creating genre: %v", err)
	}

	// Test getting existing genre
	retrieved, err := repo.GetByID(ctx, genre.ID)
	if err != nil {
		t.Errorf("Unexpected error getting genre: %v", err)
	}
	if retrieved.ID != genre.ID || retrieved.Name != genre.Name {
		t.Errorf("Retrieved genre doesn't match. Expected: %+v, Got: %+v", genre, retrieved)
	}
}

func TestGenreRepository_GetByName(t *testing.T) {
	repo := NewMockGenreRepository()
	ctx := context.Background()

	// Test getting non-existent genre by name
	_, err := repo.GetByName(ctx, "Action")
	if err == nil {
		t.Error("Expected error for non-existent genre")
	}

	// Create a genre
	genre := &models.Genre{Name: "Action"}
	err = repo.Create(ctx, genre)
	if err != nil {
		t.Errorf("Unexpected error creating genre: %v", err)
	}

	// Test getting existing genre by name
	retrieved, err := repo.GetByName(ctx, "Action")
	if err != nil {
		t.Errorf("Unexpected error getting genre by name: %v", err)
	}
	if retrieved.ID != genre.ID || retrieved.Name != genre.Name {
		t.Errorf("Retrieved genre doesn't match. Expected: %+v, Got: %+v", genre, retrieved)
	}
}

func TestGenreRepository_Create(t *testing.T) {
	repo := NewMockGenreRepository()
	ctx := context.Background()

	genre := &models.Genre{Name: "Comedy"}
	err := repo.Create(ctx, genre)
	if err != nil {
		t.Errorf("Unexpected error creating genre: %v", err)
	}

	if genre.ID == 0 {
		t.Error("Expected genre ID to be set")
	}
	if genre.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, genre.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving created genre: %v", err)
	}
	if retrieved.Name != genre.Name {
		t.Errorf("Genre name mismatch. Expected: %s, Got: %s", genre.Name, retrieved.Name)
	}
}

func TestGenreRepository_Update(t *testing.T) {
	repo := NewMockGenreRepository()
	ctx := context.Background()

	// Create a genre
	genre := &models.Genre{Name: "Drama"}
	err := repo.Create(ctx, genre)
	if err != nil {
		t.Errorf("Unexpected error creating genre: %v", err)
	}

	// Update the genre
	genre.Name = "Thriller"
	err = repo.Update(ctx, genre)
	if err != nil {
		t.Errorf("Unexpected error updating genre: %v", err)
	}

	// Verify the update
	retrieved, err := repo.GetByID(ctx, genre.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving updated genre: %v", err)
	}
	if retrieved.Name != "Thriller" {
		t.Errorf("Genre name not updated. Expected: Thriller, Got: %s", retrieved.Name)
	}

	// Test updating non-existent genre
	nonExistent := &models.Genre{ID: 999, Name: "Test"}
	err = repo.Update(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error updating non-existent genre")
	}
}

func TestGenreRepository_Delete(t *testing.T) {
	repo := NewMockGenreRepository()
	ctx := context.Background()

	// Create a genre
	genre := &models.Genre{Name: "Horror"}
	err := repo.Create(ctx, genre)
	if err != nil {
		t.Errorf("Unexpected error creating genre: %v", err)
	}

	// Delete the genre
	err = repo.Delete(ctx, genre.ID)
	if err != nil {
		t.Errorf("Unexpected error deleting genre: %v", err)
	}

	// Verify it was deleted
	_, err = repo.GetByID(ctx, genre.ID)
	if err == nil {
		t.Error("Expected error retrieving deleted genre")
	}

	// Test deleting non-existent genre
	err = repo.Delete(ctx, 999)
	if err == nil {
		t.Error("Expected error deleting non-existent genre")
	}
}

func TestGenreRepository_GetAll(t *testing.T) {
	repo := NewMockGenreRepository()
	ctx := context.Background()

	// Test empty repository
	genres, err := repo.GetAll(ctx)
	if err != nil {
		t.Errorf("Unexpected error getting all genres: %v", err)
	}
	if len(genres) != 0 {
		t.Errorf("Expected empty slice, got %d genres", len(genres))
	}

	// Create some genres
	genre1 := &models.Genre{Name: "Action"}
	genre2 := &models.Genre{Name: "Comedy"}
	err = repo.Create(ctx, genre1)
	if err != nil {
		t.Errorf("Unexpected error creating genre1: %v", err)
	}
	err = repo.Create(ctx, genre2)
	if err != nil {
		t.Errorf("Unexpected error creating genre2: %v", err)
	}

	// Test getting all genres
	genres, err = repo.GetAll(ctx)
	if err != nil {
		t.Errorf("Unexpected error getting all genres: %v", err)
	}
	if len(genres) != 2 {
		t.Errorf("Expected 2 genres, got %d", len(genres))
	}
}

func TestGenreRepository_Count(t *testing.T) {
	repo := NewMockGenreRepository()
	ctx := context.Background()

	// Test empty repository
	count, err := repo.Count(ctx)
	if err != nil {
		t.Errorf("Unexpected error counting genres: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Create some genres
	genre1 := &models.Genre{Name: "Action"}
	genre2 := &models.Genre{Name: "Comedy"}
	err = repo.Create(ctx, genre1)
	if err != nil {
		t.Errorf("Unexpected error creating genre1: %v", err)
	}
	err = repo.Create(ctx, genre2)
	if err != nil {
		t.Errorf("Unexpected error creating genre2: %v", err)
	}

	// Test counting genres
	count, err = repo.Count(ctx)
	if err != nil {
		t.Errorf("Unexpected error counting genres: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

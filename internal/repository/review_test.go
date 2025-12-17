package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"golang-project/internal/models"
)

// MockReviewRepository implements ReviewRepository for testing
type MockReviewRepository struct {
	reviews map[int]*models.Review
	nextID  int
}

func NewMockReviewRepository() *MockReviewRepository {
	return &MockReviewRepository{
		reviews: make(map[int]*models.Review),
		nextID:  1,
	}
}

func (r *MockReviewRepository) GetByID(ctx context.Context, id int) (*models.Review, error) {
	if review, exists := r.reviews[id]; exists {
		return review, nil
	}
	return nil, sql.ErrNoRows
}

func (r *MockReviewRepository) GetByMovieID(ctx context.Context, movieID int, filters models.ReviewFilters, limit, offset int) ([]models.Review, error) {
	result := make([]models.Review, 0)
	for _, review := range r.reviews {
		if review.MovieID == movieID && r.matchesFilters(review, filters) {
			result = append(result, *review)
		}
	}

	// Apply pagination
	if offset >= len(result) {
		return []models.Review{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

func (r *MockReviewRepository) GetByUserID(ctx context.Context, userID int, filters models.ReviewFilters, limit, offset int) ([]models.Review, error) {
	result := make([]models.Review, 0)
	for _, review := range r.reviews {
		if review.UserID == userID && r.matchesFilters(review, filters) {
			result = append(result, *review)
		}
	}

	// Apply pagination
	if offset >= len(result) {
		return []models.Review{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

func (r *MockReviewRepository) GetByMovieAndUser(ctx context.Context, movieID, userID int) (*models.Review, error) {
	for _, review := range r.reviews {
		if review.MovieID == movieID && review.UserID == userID {
			return review, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *MockReviewRepository) Create(ctx context.Context, review *models.Review) error {
	review.ID = r.nextID
	review.CreatedAt = time.Now()
	review.UpdatedAt = review.CreatedAt
	r.reviews[review.ID] = review
	r.nextID++
	return nil
}

func (r *MockReviewRepository) Update(ctx context.Context, review *models.Review) error {
	if _, exists := r.reviews[review.ID]; !exists {
		return sql.ErrNoRows
	}
	review.UpdatedAt = time.Now()
	r.reviews[review.ID] = review
	return nil
}

func (r *MockReviewRepository) Delete(ctx context.Context, id int) error {
	if _, exists := r.reviews[id]; !exists {
		return sql.ErrNoRows
	}
	delete(r.reviews, id)
	return nil
}

func (r *MockReviewRepository) CountByMovieID(ctx context.Context, movieID int) (int, error) {
	count := 0
	for _, review := range r.reviews {
		if review.MovieID == movieID {
			count++
		}
	}
	return count, nil
}

func (r *MockReviewRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	count := 0
	for _, review := range r.reviews {
		if review.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (r *MockReviewRepository) matchesFilters(review *models.Review, filters models.ReviewFilters) bool {
	// Check minimum rating filter
	if filters.MinRating > 0 && review.Rating < filters.MinRating {
		return false
	}

	// Check maximum rating filter
	if filters.MaxRating > 0 && review.Rating > filters.MaxRating {
		return false
	}

	return true
}

func TestReviewRepository_GetByID(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	// Test getting non-existent review
	_, err := repo.GetByID(ctx, 1)
	if err == nil {
		t.Error("Expected error for non-existent review")
	}

	// Create a review
	review := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	err = repo.Create(ctx, review)
	if err != nil {
		t.Errorf("Unexpected error creating review: %v", err)
	}

	// Test getting existing review
	retrieved, err := repo.GetByID(ctx, review.ID)
	if err != nil {
		t.Errorf("Unexpected error getting review: %v", err)
	}
	if retrieved.ID != review.ID || retrieved.Title != review.Title {
		t.Errorf("Retrieved review doesn't match. Expected: %+v, Got: %+v", review, retrieved)
	}
}

func TestReviewRepository_GetByMovieID(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	// Test empty repository
	reviews, err := repo.GetByMovieID(ctx, 1, models.ReviewFilters{}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error getting reviews by movie ID: %v", err)
	}
	if len(reviews) != 0 {
		t.Errorf("Expected empty slice, got %d reviews", len(reviews))
	}

	// Create some reviews for movie 1
	review1 := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	review2 := &models.Review{
		MovieID: 1,
		UserID:  2,
		Rating:  7,
		Title:   "Good movie",
		Content: "This movie was pretty good.",
	}
	review3 := &models.Review{
		MovieID: 2,
		UserID:  1,
		Rating:  9,
		Title:   "Excellent!",
		Content: "One of the best movies I've seen.",
	}
	err = repo.Create(ctx, review1)
	if err != nil {
		t.Errorf("Unexpected error creating review1: %v", err)
	}
	err = repo.Create(ctx, review2)
	if err != nil {
		t.Errorf("Unexpected error creating review2: %v", err)
	}
	err = repo.Create(ctx, review3)
	if err != nil {
		t.Errorf("Unexpected error creating review3: %v", err)
	}

	// Test getting reviews for movie 1
	reviews, err = repo.GetByMovieID(ctx, 1, models.ReviewFilters{}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error getting reviews by movie ID: %v", err)
	}
	if len(reviews) != 2 {
		t.Errorf("Expected 2 reviews for movie 1, got %d", len(reviews))
	}

	// Test pagination
	reviews, err = repo.GetByMovieID(ctx, 1, models.ReviewFilters{}, 1, 0)
	if err != nil {
		t.Errorf("Unexpected error getting reviews by movie ID: %v", err)
	}
	if len(reviews) != 1 {
		t.Errorf("Expected 1 review with limit 1, got %d", len(reviews))
	}
}

func TestReviewRepository_GetByUserID(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	// Create some reviews
	review1 := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	review2 := &models.Review{
		MovieID: 2,
		UserID:  1,
		Rating:  7,
		Title:   "Good movie",
		Content: "This movie was pretty good.",
	}
	review3 := &models.Review{
		MovieID: 1,
		UserID:  2,
		Rating:  9,
		Title:   "Excellent!",
		Content: "One of the best movies I've seen.",
	}
	err := repo.Create(ctx, review1)
	if err != nil {
		t.Errorf("Unexpected error creating review1: %v", err)
	}
	err = repo.Create(ctx, review2)
	if err != nil {
		t.Errorf("Unexpected error creating review2: %v", err)
	}
	err = repo.Create(ctx, review3)
	if err != nil {
		t.Errorf("Unexpected error creating review3: %v", err)
	}

	// Test getting reviews for user 1
	reviews, err := repo.GetByUserID(ctx, 1, models.ReviewFilters{}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error getting reviews by user ID: %v", err)
	}
	if len(reviews) != 2 {
		t.Errorf("Expected 2 reviews for user 1, got %d", len(reviews))
	}

	// Test getting reviews for user 2
	reviews, err = repo.GetByUserID(ctx, 2, models.ReviewFilters{}, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error getting reviews by user ID: %v", err)
	}
	if len(reviews) != 1 {
		t.Errorf("Expected 1 review for user 2, got %d", len(reviews))
	}
}

func TestReviewRepository_GetByMovieAndUser(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	// Test getting non-existent review
	_, err := repo.GetByMovieAndUser(ctx, 1, 1)
	if err == nil {
		t.Error("Expected error for non-existent review")
	}

	// Create a review
	review := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	err = repo.Create(ctx, review)
	if err != nil {
		t.Errorf("Unexpected error creating review: %v", err)
	}

	// Test getting existing review
	retrieved, err := repo.GetByMovieAndUser(ctx, 1, 1)
	if err != nil {
		t.Errorf("Unexpected error getting review by movie and user: %v", err)
	}
	if retrieved.ID != review.ID || retrieved.Title != review.Title {
		t.Errorf("Retrieved review doesn't match. Expected: %+v, Got: %+v", review, retrieved)
	}

	// Test getting non-existent review for different user
	_, err = repo.GetByMovieAndUser(ctx, 1, 2)
	if err == nil {
		t.Error("Expected error for non-existent review")
	}
}

func TestReviewRepository_Create(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	review := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	err := repo.Create(ctx, review)
	if err != nil {
		t.Errorf("Unexpected error creating review: %v", err)
	}

	if review.ID == 0 {
		t.Error("Expected review ID to be set")
	}
	if review.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if review.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, review.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving created review: %v", err)
	}
	if retrieved.Title != review.Title {
		t.Errorf("Review title mismatch. Expected: %s, Got: %s", review.Title, retrieved.Title)
	}
}

func TestReviewRepository_Update(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	// Create a review
	review := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	err := repo.Create(ctx, review)
	if err != nil {
		t.Errorf("Unexpected error creating review: %v", err)
	}

	// Update the review
	originalUpdatedAt := review.UpdatedAt
	review.Rating = 9
	review.Title = "Excellent movie!"
	review.Content = "This is an excellent movie!"
	err = repo.Update(ctx, review)
	if err != nil {
		t.Errorf("Unexpected error updating review: %v", err)
	}

	// Verify the update
	retrieved, err := repo.GetByID(ctx, review.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving updated review: %v", err)
	}
	if retrieved.Rating != 9 {
		t.Errorf("Review rating not updated. Expected: 9, Got: %d", retrieved.Rating)
	}
	if retrieved.Title != "Excellent movie!" {
		t.Errorf("Review title not updated. Expected: Excellent movie!, Got: %s", retrieved.Title)
	}
	if retrieved.UpdatedAt.Before(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}

	// Test updating non-existent review
	nonExistent := &models.Review{ID: 999, Title: "Test"}
	err = repo.Update(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error updating non-existent review")
	}
}

func TestReviewRepository_Delete(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	// Create a review
	review := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	err := repo.Create(ctx, review)
	if err != nil {
		t.Errorf("Unexpected error creating review: %v", err)
	}

	// Delete the review
	err = repo.Delete(ctx, review.ID)
	if err != nil {
		t.Errorf("Unexpected error deleting review: %v", err)
	}

	// Verify it was deleted
	_, err = repo.GetByID(ctx, review.ID)
	if err == nil {
		t.Error("Expected error retrieving deleted review")
	}

	// Test deleting non-existent review
	err = repo.Delete(ctx, 999)
	if err == nil {
		t.Error("Expected error deleting non-existent review")
	}
}

func TestReviewRepository_CountByMovieID(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	// Test empty repository
	count, err := repo.CountByMovieID(ctx, 1)
	if err != nil {
		t.Errorf("Unexpected error counting reviews by movie ID: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Create some reviews
	review1 := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	review2 := &models.Review{
		MovieID: 1,
		UserID:  2,
		Rating:  7,
		Title:   "Good movie",
		Content: "This movie was pretty good.",
	}
	review3 := &models.Review{
		MovieID: 2,
		UserID:  1,
		Rating:  9,
		Title:   "Excellent!",
		Content: "One of the best movies I've seen.",
	}
	err = repo.Create(ctx, review1)
	if err != nil {
		t.Errorf("Unexpected error creating review1: %v", err)
	}
	err = repo.Create(ctx, review2)
	if err != nil {
		t.Errorf("Unexpected error creating review2: %v", err)
	}
	err = repo.Create(ctx, review3)
	if err != nil {
		t.Errorf("Unexpected error creating review3: %v", err)
	}

	// Test counting reviews for movie 1
	count, err = repo.CountByMovieID(ctx, 1)
	if err != nil {
		t.Errorf("Unexpected error counting reviews by movie ID: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2 for movie 1, got %d", count)
	}

	// Test counting reviews for movie 2
	count, err = repo.CountByMovieID(ctx, 2)
	if err != nil {
		t.Errorf("Unexpected error counting reviews by movie ID: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1 for movie 2, got %d", count)
	}
}

func TestReviewRepository_CountByUserID(t *testing.T) {
	repo := NewMockReviewRepository()
	ctx := context.Background()

	// Test empty repository
	count, err := repo.CountByUserID(ctx, 1)
	if err != nil {
		t.Errorf("Unexpected error counting reviews by user ID: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Create some reviews
	review1 := &models.Review{
		MovieID: 1,
		UserID:  1,
		Rating:  8,
		Title:   "Great movie!",
		Content: "I really enjoyed this movie.",
	}
	review2 := &models.Review{
		MovieID: 2,
		UserID:  1,
		Rating:  7,
		Title:   "Good movie",
		Content: "This movie was pretty good.",
	}
	review3 := &models.Review{
		MovieID: 1,
		UserID:  2,
		Rating:  9,
		Title:   "Excellent!",
		Content: "One of the best movies I've seen.",
	}
	err = repo.Create(ctx, review1)
	if err != nil {
		t.Errorf("Unexpected error creating review1: %v", err)
	}
	err = repo.Create(ctx, review2)
	if err != nil {
		t.Errorf("Unexpected error creating review2: %v", err)
	}
	err = repo.Create(ctx, review3)
	if err != nil {
		t.Errorf("Unexpected error creating review3: %v", err)
	}

	// Test counting reviews for user 1
	count, err = repo.CountByUserID(ctx, 1)
	if err != nil {
		t.Errorf("Unexpected error counting reviews by user ID: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2 for user 1, got %d", count)
	}

	// Test counting reviews for user 2
	count, err = repo.CountByUserID(ctx, 2)
	if err != nil {
		t.Errorf("Unexpected error counting reviews by user ID: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1 for user 2, got %d", count)
	}
}

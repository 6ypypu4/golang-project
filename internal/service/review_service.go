package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"golang-project/internal/models"
)

var (
	ErrReviewExists   = errors.New("review already exists")
	ErrReviewNotFound = errors.New("review not found")
)

type ReviewRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Review, error)
	GetByMovieAndUser(ctx context.Context, movieID, userID uuid.UUID) (*models.Review, error)
	GetByMovieID(ctx context.Context, movieID uuid.UUID, limit, offset int) ([]models.Review, error)
	Create(ctx context.Context, review *models.Review) error
	Update(ctx context.Context, review *models.Review) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type MovieLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Movie, error)
	UpdateAverageRating(ctx context.Context, movieID uuid.UUID) error
}

type ReviewService struct {
	reviews   ReviewRepo
	movies    MovieLookup
	validator *validator.Validate
}

func NewReviewService(reviews ReviewRepo, movies MovieLookup, v *validator.Validate) *ReviewService {
	return &ReviewService{
		reviews:   reviews,
		movies:    movies,
		validator: v,
	}
}

func (s *ReviewService) ListByMovie(ctx context.Context, movieID uuid.UUID, page, limit int) ([]models.Review, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.reviews.GetByMovieID(ctx, movieID, limit, offset)
}

func (s *ReviewService) Create(ctx context.Context, movieID, userID uuid.UUID, req models.CreateReviewRequest) (*models.Review, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, err
	}

	// ensure movie exists
	if _, err := s.movies.GetByID(ctx, movieID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMovieNotFound
		}
		return nil, err
	}

	if _, err := s.reviews.GetByMovieAndUser(ctx, movieID, userID); err == nil {
		return nil, ErrReviewExists
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	review := &models.Review{
		MovieID: movieID,
		UserID:  userID,
		Rating:  req.Rating,
		Title:   req.Title,
		Content: req.Content,
	}
	if err := s.reviews.Create(ctx, review); err != nil {
		return nil, err
	}
	_ = s.movies.UpdateAverageRating(ctx, movieID)
	return review, nil
}

func (s *ReviewService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req models.UpdateReviewRequest) (*models.Review, error) {
	review, err := s.reviews.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	if review.UserID != userID {
		return nil, ErrInvalidCredentials
	}

	if req.Rating != 0 {
		review.Rating = req.Rating
	}
	if req.Title != "" {
		review.Title = req.Title
	}
	if req.Content != "" {
		review.Content = req.Content
	}

	if err := s.reviews.Update(ctx, review); err != nil {
		return nil, err
	}
	_ = s.movies.UpdateAverageRating(ctx, review.MovieID)
	return review, nil
}

func (s *ReviewService) Delete(ctx context.Context, id uuid.UUID, requester uuid.UUID, isAdmin bool) error {
	review, err := s.reviews.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrReviewNotFound
		}
		return err
	}
	if !isAdmin && review.UserID != requester {
		return ErrInvalidCredentials
	}

	if err := s.reviews.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.movies.UpdateAverageRating(ctx, review.MovieID)
	return nil
}

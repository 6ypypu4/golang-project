package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-playground/validator/v10"

	"golang-project/internal/models"
)

var (
	ErrGenreExists   = errors.New("genre already exists")
	ErrGenreNotFound = errors.New("genre not found")
)

type GenreRepo interface {
	GetAll(ctx context.Context) ([]models.Genre, error)
	GetByID(ctx context.Context, id int) (*models.Genre, error)
	GetByName(ctx context.Context, name string) (*models.Genre, error)
	Create(ctx context.Context, genre *models.Genre) error
	Update(ctx context.Context, genre *models.Genre) error
	Delete(ctx context.Context, id int) error
}

type GenreService struct {
	repo      GenreRepo
	validator *validator.Validate
}

func NewGenreService(repo GenreRepo, v *validator.Validate) *GenreService {
	return &GenreService{repo: repo, validator: v}
}

func (s *GenreService) List(ctx context.Context) ([]models.Genre, error) {
	return s.repo.GetAll(ctx)
}

func (s *GenreService) Get(ctx context.Context, id int) (*models.Genre, error) {
	genre, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGenreNotFound
		}
		return nil, err
	}
	return genre, nil
}

func (s *GenreService) Create(ctx context.Context, req models.CreateGenreRequest) (*models.Genre, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetByName(ctx, req.Name); err == nil {
		return nil, ErrGenreExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	genre := &models.Genre{Name: req.Name}
	if err := s.repo.Create(ctx, genre); err != nil {
		return nil, err
	}
	return genre, nil
}

func (s *GenreService) Update(ctx context.Context, id int, req models.CreateGenreRequest) (*models.Genre, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, err
	}

	genre, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGenreNotFound
		}
		return nil, err
	}

	genre.Name = req.Name
	if err := s.repo.Update(ctx, genre); err != nil {
		return nil, err
	}
	return genre, nil
}

func (s *GenreService) Delete(ctx context.Context, id int) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGenreNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

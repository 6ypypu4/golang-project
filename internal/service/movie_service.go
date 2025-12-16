package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"

	"golang-project/internal/models"
)

var (
	ErrMovieNotFound    = errors.New("movie not found")
	ErrNoGenresProvided = errors.New("at least one genre required")
)

type MovieRepo interface {
	List(ctx context.Context, filters models.MovieFilters, limit, offset int) ([]models.Movie, int, error)
	GetByID(ctx context.Context, id int) (*models.Movie, error)
	Create(ctx context.Context, movie *models.Movie) error
	Update(ctx context.Context, movie *models.Movie) error
	Delete(ctx context.Context, id int) error
	SetGenres(ctx context.Context, movieID int, genreIDs []int) error
	GetGenresByMovieID(ctx context.Context, movieID int) ([]models.Genre, error)
}

type GenreLookup interface {
	GetByID(ctx context.Context, id int) (*models.Genre, error)
}

type MovieService struct {
	movies    MovieRepo
	genres    GenreLookup
	validator *validator.Validate
}

func NewMovieService(movies MovieRepo, genres GenreLookup, v *validator.Validate) *MovieService {
	return &MovieService{
		movies:    movies,
		genres:    genres,
		validator: v,
	}
}

func (s *MovieService) List(ctx context.Context, filters models.MovieFilters, page, limit int) (*models.PaginatedResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	movies, total, err := s.movies.List(ctx, filters, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	return &models.PaginatedResponse{
		Data:       movies,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *MovieService) Get(ctx context.Context, id int) (*models.Movie, error) {
	movie, err := s.movies.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMovieNotFound
		}
		return nil, err
	}

	genres, err := s.movies.GetGenresByMovieID(ctx, id)
	if err != nil {
		return nil, err
	}
	movie.Genres = genres
	return movie, nil
}

func (s *MovieService) Create(ctx context.Context, req models.CreateMovieRequest) (*models.Movie, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, err
	}
	if len(req.GenreIDs) == 0 {
		return nil, ErrNoGenresProvided
	}

	genreIDs, err := s.validateGenreIDs(ctx, req.GenreIDs)
	if err != nil {
		return nil, err
	}

	movie := &models.Movie{
		Title:           req.Title,
		Description:     req.Description,
		ReleaseYear:     req.ReleaseYear,
		Director:        req.Director,
		DurationMinutes: req.DurationMinutes,
	}

	if err := s.movies.Create(ctx, movie); err != nil {
		return nil, err
	}
	if err := s.movies.SetGenres(ctx, movie.ID, genreIDs); err != nil {
		return nil, err
	}
	movie.Genres = make([]models.Genre, 0, len(genreIDs))
	for _, gid := range genreIDs {
		g, err := s.genres.GetByID(ctx, gid)
		if err != nil {
			return nil, err
		}
		movie.Genres = append(movie.Genres, *g)
	}
	return movie, nil
}

func (s *MovieService) Update(ctx context.Context, id int, req models.UpdateMovieRequest) (*models.Movie, error) {
	movie, err := s.movies.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMovieNotFound
		}
		return nil, err
	}

	// Apply partial updates; validator only on provided fields
	if req.Title != "" {
		movie.Title = req.Title
	}
	if req.Description != "" {
		movie.Description = req.Description
	}
	if req.ReleaseYear != 0 {
		movie.ReleaseYear = req.ReleaseYear
	}
	if req.Director != "" {
		movie.Director = req.Director
	}
	if req.DurationMinutes != 0 {
		movie.DurationMinutes = req.DurationMinutes
	}

	if err := s.movies.Update(ctx, movie); err != nil {
		return nil, err
	}

	if req.GenreIDs != nil {
		if len(req.GenreIDs) == 0 {
			return nil, ErrNoGenresProvided
		}
		genreIDs, err := s.validateGenreIDs(ctx, req.GenreIDs)
		if err != nil {
			return nil, err
		}
		if err := s.movies.SetGenres(ctx, movie.ID, genreIDs); err != nil {
			return nil, err
		}
		movie.Genres = make([]models.Genre, 0, len(genreIDs))
		for _, gid := range genreIDs {
			g, err := s.genres.GetByID(ctx, gid)
			if err != nil {
				return nil, err
			}
			movie.Genres = append(movie.Genres, *g)
		}
	} else {
		genres, err := s.movies.GetGenresByMovieID(ctx, movie.ID)
		if err != nil {
			return nil, err
		}
		movie.Genres = genres
	}

	return movie, nil
}

func (s *MovieService) Delete(ctx context.Context, id int) error {
	if _, err := s.movies.GetByID(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrMovieNotFound
		}
		return err
	}
	return s.movies.Delete(ctx, id)
}

func (s *MovieService) validateGenreIDs(ctx context.Context, ids []string) ([]int, error) {
	genreIDs := make([]int, 0, len(ids))
	for _, idStr := range ids {
		genreID, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, ErrGenreNotFound
		}
		if _, err := s.genres.GetByID(ctx, genreID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrGenreNotFound
			}
			return nil, err
		}
		genreIDs = append(genreIDs, genreID)
	}
	return genreIDs, nil
}

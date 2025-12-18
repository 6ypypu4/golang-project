package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-playground/validator/v10"

	"golang-project/internal/models"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidRole      = errors.New("invalid role")
	ErrCannotDeleteSelf = errors.New("cannot delete yourself")
	allowedRoles        = map[string]struct{}{"user": {}, "admin": {}}
)

type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
	List(ctx context.Context, filters models.UserFilters, limit, offset int) ([]models.User, int, error)
	UpdateRole(ctx context.Context, id int, role string) error
	Update(ctx context.Context, id int, email, username string) error
	UpdatePassword(ctx context.Context, id int, passwordHash string) error
	Delete(ctx context.Context, id int) error
	Count(ctx context.Context) (int, error)
	CountLast7Days(ctx context.Context) (int, error)
}

type ReviewStatsRepo interface {
	GetAverageRatingByUserID(ctx context.Context, userID int) (float64, error)
	GetFavoriteGenreByUserID(ctx context.Context, userID int) (*models.Genre, error)
}

type UserService struct {
	repo           UserRepo
	reviewStats    ReviewStatsRepo
	validator      *validator.Validate
	passwordHasher PasswordHasher
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(hash, password string) error
}

func NewUserService(repo UserRepo, reviewStats ReviewStatsRepo, v *validator.Validate, passwordHasher PasswordHasher) *UserService {
	return &UserService{
		repo:           repo,
		reviewStats:    reviewStats,
		validator:      v,
		passwordHasher: passwordHasher,
	}
}

func (s *UserService) GetByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) List(ctx context.Context, filters models.UserFilters, page, limit int) (*models.PaginatedResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit
	users, total, err := s.repo.List(ctx, filters, limit, offset)
	if err != nil {
		return nil, err
	}
	totalPages := (total + limit - 1) / limit
	return &models.PaginatedResponse{
		Data:       users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *UserService) UpdateRole(ctx context.Context, id int, role string) error {
	if _, ok := allowedRoles[role]; !ok {
		return ErrInvalidRole
	}
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrUserNotFound
	}
	return s.repo.UpdateRole(ctx, id, role)
}

func (s *UserService) Update(ctx context.Context, id int, req models.UpdateUserRequest) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if err := s.validator.Struct(req); err != nil {
		return err
	}

	email := req.Email
	username := req.Username

	if email != "" && email != user.Email {
		existing, err := s.repo.GetByEmail(ctx, email)
		if err == nil && existing.ID != id {
			return ErrUserExists
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	if username != "" && username != user.Username {
		existing, err := s.repo.GetByUsername(ctx, username)
		if err == nil && existing.ID != id {
			return ErrUserExists
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	return s.repo.Update(ctx, id, email, username)
}

func (s *UserService) Delete(ctx context.Context, id int, adminID int) error {
	if id == adminID {
		return ErrCannotDeleteSelf
	}

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int, req models.UpdateUserRequest) (*models.User, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, err
	}

	if err := s.Update(ctx, userID, req); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, userID)
}

func (s *UserService) UpdatePassword(ctx context.Context, userID int, currentPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if err := s.passwordHasher.CheckPassword(user.PasswordHash, currentPassword); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := s.passwordHasher.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, hash)
}

func (s *UserService) GetUserStats(ctx context.Context, userID int) (*models.UserStats, error) {
	avgRating, err := s.reviewStats.GetAverageRatingByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	favoriteGenre, err := s.reviewStats.GetFavoriteGenreByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats := &models.UserStats{
		AverageRating: avgRating,
	}
	if favoriteGenre != nil {
		stats.FavoriteGenre = favoriteGenre
	}

	return stats, nil
}

func (s *UserService) GetAdminStats(ctx context.Context, userRepo UserRepo, movieRepo MovieCountRepo, reviewRepo ReviewCountRepo, genreRepo GenreCountRepo) (*models.AdminStats, error) {
	totalUsers, err := userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalMovies, err := movieRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalReviews, err := reviewRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalGenres, err := genreRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	avgRating, err := movieRepo.GetAverageRating(ctx)
	if err != nil {
		return nil, err
	}

	usersLast7Days, err := userRepo.CountLast7Days(ctx)
	if err != nil {
		return nil, err
	}

	reviewsLast7Days, err := reviewRepo.CountLast7Days(ctx)
	if err != nil {
		return nil, err
	}

	moviesLast7Days, err := movieRepo.CountLast7Days(ctx)
	if err != nil {
		return nil, err
	}

	return &models.AdminStats{
		TotalUsers:       totalUsers,
		TotalMovies:      totalMovies,
		TotalReviews:     totalReviews,
		TotalGenres:      totalGenres,
		AverageRating:    avgRating,
		UsersLast7Days:   usersLast7Days,
		ReviewsLast7Days: reviewsLast7Days,
		MoviesLast7Days:  moviesLast7Days,
	}, nil
}

func (s *UserService) ListAuditLogs(ctx context.Context, auditRepo AuditLogRepo, filters models.AuditLogFilters, page, limit int) (*models.PaginatedResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	logs, total, err := auditRepo.List(ctx, filters, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	return &models.PaginatedResponse{
		Data:       logs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

type MovieCountRepo interface {
	Count(ctx context.Context) (int, error)
	GetAverageRating(ctx context.Context) (float64, error)
	CountLast7Days(ctx context.Context) (int, error)
}

type ReviewCountRepo interface {
	Count(ctx context.Context) (int, error)
	CountLast7Days(ctx context.Context) (int, error)
}

type GenreCountRepo interface {
	Count(ctx context.Context) (int, error)
}

type AuditLogRepo interface {
	List(ctx context.Context, filters models.AuditLogFilters, limit, offset int) ([]models.AuditLog, int, error)
}

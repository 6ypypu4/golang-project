package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-playground/validator/v10"

	"golang-project/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidRole  = errors.New("invalid role")
	allowedRoles    = map[string]struct{}{"user": {}, "admin": {}}
)

type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
	List(ctx context.Context, limit, offset int) ([]models.User, int, error)
	UpdateRole(ctx context.Context, id int, role string) error
}

type UserService struct {
	repo      UserRepo
	validator *validator.Validate
}

func NewUserService(repo UserRepo, v *validator.Validate) *UserService {
	return &UserService{repo: repo, validator: v}
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

func (s *UserService) List(ctx context.Context, page, limit int) (*models.PaginatedResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit
	users, total, err := s.repo.List(ctx, limit, offset)
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

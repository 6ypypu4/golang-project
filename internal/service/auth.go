package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"

	"golang-project/internal/models"
	"golang-project/internal/repository"
	"golang-project/pkg/jwt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	users     repository.UserRepository
	validator *validator.Validate
	jwtSecret string
	tokenTTL  time.Duration
}

func NewAuthService(users repository.UserRepository, validator *validator.Validate, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		validator: validator,
		jwtSecret: jwtSecret,
		tokenTTL:  24 * time.Hour,
	}
}

func (s *AuthService) Register(ctx context.Context, req models.CreateUserRequest) (*models.User, string, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, "", err
	}

	if _, err := s.users.GetByEmail(ctx, req.Email); err == nil {
		return nil, "", ErrUserExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, "", err
	}

	hash, err := jwt.HashPassword(req.Password)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hash,
		Role:         "user",
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := jwt.Generate(fmt.Sprintf("%d", user.ID), user.Role, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.User, string, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, "", err
	}

	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := jwt.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := jwt.Generate(fmt.Sprintf("%d", user.ID), user.Role, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

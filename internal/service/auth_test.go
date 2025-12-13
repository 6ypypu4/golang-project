package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"golang-project/internal/models"
	"golang-project/pkg/jwt"
)

type memoryUserRepo struct {
	users map[string]*models.User
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{users: make(map[string]*models.User)}
}

func (r *memoryUserRepo) Create(ctx context.Context, user *models.User) error {
	if _, ok := r.users[user.Email]; ok {
		return errors.New("duplicate")
	}
	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now
	r.users[user.Email] = user
	return nil
}

func (r *memoryUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if u, ok := r.users[email]; ok {
		return u, nil
	}
	return nil, sql.ErrNoRows
}

func TestAuthService_Register(t *testing.T) {
	secret := "test-secret"
	repo := newMemoryUserRepo()
	svc := NewAuthService(repo, validator.New(), secret)

	t.Run("ok", func(t *testing.T) {
		req := models.CreateUserRequest{
			Email:    "user@example.com",
			Username: "user",
			Password: "password123",
		}
		user, token, err := svc.Register(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user == nil || user.ID == uuid.Nil {
			t.Fatalf("expected user with ID, got %#v", user)
		}
		if token == "" {
			t.Fatalf("expected token, got empty")
		}
		if user.Role != "user" {
			t.Fatalf("expected role user, got %s", user.Role)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		existing := &models.User{
			ID:           uuid.New(),
			Email:        "dupe@example.com",
			Username:     "dupe",
			PasswordHash: "hash",
			Role:         "user",
		}
		repo.users[existing.Email] = existing

		req := models.CreateUserRequest{
			Email:    existing.Email,
			Username: "newuser",
			Password: "password123",
		}
		if _, _, err := svc.Register(context.Background(), req); err == nil || !errors.Is(err, ErrUserExists) {
			t.Fatalf("expected ErrUserExists, got %v", err)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		req := models.CreateUserRequest{
			Email:    "bad-email",
			Username: "us",
			Password: "123",
		}
		if _, _, err := svc.Register(context.Background(), req); err == nil {
			t.Fatalf("expected validation error")
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	secret := "test-secret"
	repo := newMemoryUserRepo()
	svc := NewAuthService(repo, validator.New(), secret)

	password := "password123"
	hash, err := jwt.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Username:     "user",
		PasswordHash: hash,
		Role:         "user",
	}
	repo.users[user.Email] = user

	t.Run("ok", func(t *testing.T) {
		req := models.LoginRequest{
			Email:    user.Email,
			Password: password,
		}
		u, token, err := svc.Login(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u == nil || u.ID != user.ID {
			t.Fatalf("expected user %s, got %#v", user.ID, u)
		}
		if token == "" {
			t.Fatalf("expected token, got empty")
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		req := models.LoginRequest{
			Email:    user.Email,
			Password: "wrong",
		}
		if _, _, err := svc.Login(context.Background(), req); err == nil || !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		req := models.LoginRequest{
			Email:    "missing@example.com",
			Password: password,
		}
		if _, _, err := svc.Login(context.Background(), req); err == nil || !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

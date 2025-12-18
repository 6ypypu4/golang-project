package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

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
	user.ID = len(r.users) + 1
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

func (r *memoryUserRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memoryUserRepo) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	result := make([]models.User, 0, len(r.users))
	for _, u := range r.users {
		result = append(result, *u)
	}
	total := len(result)
	if offset >= total {
		return []models.User{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return result[offset:end], total, nil
}

func (r *memoryUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memoryUserRepo) UpdateRole(ctx context.Context, id int, role string) error {
	for _, u := range r.users {
		if u.ID == id {
			u.Role = role
			return nil
		}
	}
	return sql.ErrNoRows
}

func (r *memoryUserRepo) Update(ctx context.Context, id int, email, username string) error {
	for _, u := range r.users {
		if u.ID == id {
			if email != "" {
				delete(r.users, u.Email)
				u.Email = email
				r.users[email] = u
			}
			if username != "" {
				u.Username = username
			}
			u.UpdatedAt = time.Now()
			return nil
		}
	}
	return sql.ErrNoRows
}

func (r *memoryUserRepo) Delete(ctx context.Context, id int) error {
	for email, u := range r.users {
		if u.ID == id {
			delete(r.users, email)
			return nil
		}
	}
	return sql.ErrNoRows
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
		if user == nil || user.ID == 0 {
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
			ID:           1,
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
		ID:           1,
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
			t.Fatalf("expected user %d, got %#v", user.ID, u)
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

package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"golang-project/internal/models"
	"golang-project/internal/service"
	"golang-project/pkg/jwt"
)

type memoryUserRepo struct {
	mu    sync.Mutex
	users map[string]*models.User
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{users: make(map[string]*models.User)}
}

func (r *memoryUserRepo) Create(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
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
	r.mu.Lock()
	defer r.mu.Unlock()

	if u, ok := r.users[email]; ok {
		return u, nil
	}
	return nil, sql.ErrNoRows
}

func (r *memoryUserRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memoryUserRepo) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memoryUserRepo) UpdateRole(ctx context.Context, id int, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.ID == id {
			u.Role = role
			return nil
		}
	}
	return sql.ErrNoRows
}

func (r *memoryUserRepo) Update(ctx context.Context, id int, email, username string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for emailKey, u := range r.users {
		if u.ID == id {
			if email != "" && email != u.Email {
				delete(r.users, emailKey)
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
	r.mu.Lock()
	defer r.mu.Unlock()
	for email, u := range r.users {
		if u.ID == id {
			delete(r.users, email)
			return nil
		}
	}
	return sql.ErrNoRows
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		body        interface{}
		prepopulate bool
		wantStatus  int
		expectedErr error
	}{
		{
			name:       "success",
			body:       models.CreateUserRequest{Email: "user@example.com", Username: "user", Password: "password123"},
			wantStatus: http.StatusCreated,
		},
		{
			name:        "duplicate email",
			body:        models.CreateUserRequest{Email: "dupe@example.com", Username: "user2", Password: "password123"},
			prepopulate: true,
			wantStatus:  http.StatusConflict,
		},
		{
			name:       "validation error",
			body:       models.CreateUserRequest{Email: "bad", Username: "u", Password: "123"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad json",
			body:       `{"email": 123}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMemoryUserRepo()
			v := validator.New()
			authService := service.NewAuthService(repo, v, "secret")
			h := NewAuthHandler(authService)

			if tt.prepopulate {
				hash, _ := jwt.HashPassword("password123")
				repo.users["dupe@example.com"] = &models.User{
					ID:           1,
					Email:        "dupe@example.com",
					Username:     "dupe",
					PasswordHash: hash,
					Role:         "user",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			}

			r := gin.New()
			r.POST("/auth/register", h.Register)

			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				if s, ok := tt.body.(string); ok {
					bodyBytes = []byte(s)
				} else {
					t.Fatalf("marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		setupUser  bool
		password   string
		wantStatus int
	}{
		{
			name:       "success",
			body:       models.LoginRequest{Email: "user@example.com", Password: "password123"},
			setupUser:  true,
			password:   "password123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid password",
			body:       models.LoginRequest{Email: "user@example.com", Password: "wrong"},
			setupUser:  true,
			password:   "password123",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "user not found",
			body:       models.LoginRequest{Email: "missing@example.com", Password: "password123"},
			setupUser:  false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "validation error",
			body:       models.LoginRequest{Email: "bad", Password: ""},
			setupUser:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad json",
			body:       `{"email": 123}`,
			setupUser:  false,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMemoryUserRepo()
			if tt.setupUser {
				hash, _ := jwt.HashPassword(tt.password)
				repo.users["user@example.com"] = &models.User{
					ID:           1,
					Email:        "user@example.com",
					Username:     "user",
					PasswordHash: hash,
					Role:         "user",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
			}

			v := validator.New()
			authService := service.NewAuthService(repo, v, "secret")
			h := NewAuthHandler(authService)

			r := gin.New()
			r.POST("/auth/login", h.Login)

			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				if s, ok := tt.body.(string); ok {
					bodyBytes = []byte(s)
				} else {
					t.Fatalf("marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

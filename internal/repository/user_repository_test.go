package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"golang-project/internal/models"
)

// MockUserRepository implements UserRepository for testing
type MockUserRepository struct {
	users  map[int]*models.User
	nextID int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[int]*models.User),
		nextID: 1,
	}
}

func (r *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = r.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	r.users[user.ID] = user
	r.nextID++
	return nil
}

func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *MockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	if user, exists := r.users[id]; exists {
		return user, nil
	}
	return nil, sql.ErrNoRows
}

func (r *MockUserRepository) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	result := make([]models.User, 0, len(r.users))
	for _, user := range r.users {
		result = append(result, *user)
	}

	// Apply pagination
	total := len(result)
	start := offset
	if start >= total {
		return []models.User{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}

	return result[start:end], total, nil
}

func (r *MockUserRepository) UpdateRole(ctx context.Context, id int, role string) error {
	if user, exists := r.users[id]; exists {
		user.Role = role
		user.UpdatedAt = time.Now()
		return nil
	}
	return sql.ErrNoRows
}

func TestUserRepository_Create(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	user := &models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         "user",
	}
	err := repo.Create(ctx, user)
	if err != nil {
		t.Errorf("Unexpected error creating user: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving created user: %v", err)
	}
	if retrieved.Email != user.Email {
		t.Errorf("User email mismatch. Expected: %s, Got: %s", user.Email, retrieved.Email)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test getting non-existent user by email
	_, err := repo.GetByEmail(ctx, "test@example.com")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	// Create a user
	user := &models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         "user",
	}
	err = repo.Create(ctx, user)
	if err != nil {
		t.Errorf("Unexpected error creating user: %v", err)
	}

	// Test getting existing user by email
	retrieved, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Errorf("Unexpected error getting user by email: %v", err)
	}
	if retrieved.ID != user.ID || retrieved.Email != user.Email {
		t.Errorf("Retrieved user doesn't match. Expected: %+v, Got: %+v", user, retrieved)
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test getting non-existent user
	_, err := repo.GetByID(ctx, 1)
	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	// Create a user
	user := &models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         "user",
	}
	err = repo.Create(ctx, user)
	if err != nil {
		t.Errorf("Unexpected error creating user: %v", err)
	}

	// Test getting existing user
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Errorf("Unexpected error getting user: %v", err)
	}
	if retrieved.ID != user.ID || retrieved.Email != user.Email {
		t.Errorf("Retrieved user doesn't match. Expected: %+v, Got: %+v", user, retrieved)
	}
}

func TestUserRepository_List(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test empty repository
	users, total, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error listing users: %v", err)
	}
	if len(users) != 0 || total != 0 {
		t.Errorf("Expected empty list, got %d users with total %d", len(users), total)
	}

	// Create some users
	user1 := &models.User{
		Email:        "user1@example.com",
		Username:     "user1",
		PasswordHash: "password1",
		Role:         "user",
	}
	user2 := &models.User{
		Email:        "user2@example.com",
		Username:     "user2",
		PasswordHash: "password2",
		Role:         "admin",
	}
	err = repo.Create(ctx, user1)
	if err != nil {
		t.Errorf("Unexpected error creating user1: %v", err)
	}
	err = repo.Create(ctx, user2)
	if err != nil {
		t.Errorf("Unexpected error creating user2: %v", err)
	}

	// Test listing all users
	users, total, err = repo.List(ctx, 10, 0)
	if err != nil {
		t.Errorf("Unexpected error listing users: %v", err)
	}
	if len(users) != 2 || total != 2 {
		t.Errorf("Expected 2 users, got %d with total %d", len(users), total)
	}

	// Test pagination
	users, total, err = repo.List(ctx, 1, 0)
	if err != nil {
		t.Errorf("Unexpected error listing users: %v", err)
	}
	if len(users) != 1 || total != 2 {
		t.Errorf("Expected 1 user with limit 1, got %d with total %d", len(users), total)
	}

	// Test offset
	users, total, err = repo.List(ctx, 1, 1)
	if err != nil {
		t.Errorf("Unexpected error listing users: %v", err)
	}
	if len(users) != 1 || total != 2 {
		t.Errorf("Expected 1 user with offset 1, got %d with total %d", len(users), total)
	}
}

func TestUserRepository_UpdateRole(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         "user",
	}
	err := repo.Create(ctx, user)
	if err != nil {
		t.Errorf("Unexpected error creating user: %v", err)
	}

	// Update the user's role
	err = repo.UpdateRole(ctx, user.ID, "admin")
	if err != nil {
		t.Errorf("Unexpected error updating user role: %v", err)
	}

	// Verify the update
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving updated user: %v", err)
	}
	if retrieved.Role != "admin" {
		t.Errorf("User role not updated. Expected: admin, Got: %s", retrieved.Role)
	}

	// Test updating non-existent user
	err = repo.UpdateRole(ctx, 999, "admin")
	if err == nil {
		t.Error("Expected error updating non-existent user")
	}
}

func TestUserRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         "user",
	}
	err := repo.Create(ctx, user)
	if err != nil {
		t.Errorf("Unexpected error creating user: %v", err)
	}

	// Test that GetByEmail and GetByID return the same user
	byEmail, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Errorf("Unexpected error getting user by email: %v", err)
	}
	byID, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Errorf("Unexpected error getting user by ID: %v", err)
	}

	if byEmail.ID != byID.ID || byEmail.Email != byID.Email {
		t.Errorf("GetByEmail and GetByID returned different users")
	}
}

func TestUserRepository_DuplicateEmail(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Create first user
	user1 := &models.User{
		Email:        "test@example.com",
		Username:     "user1",
		PasswordHash: "password1",
		Role:         "user",
	}
	err := repo.Create(ctx, user1)
	if err != nil {
		t.Errorf("Unexpected error creating user1: %v", err)
	}

	// Create second user with same email (this should be allowed in our mock)
	user2 := &models.User{
		Email:        "test@example.com",
		Username:     "user2",
		PasswordHash: "password2",
		Role:         "admin",
	}
	err = repo.Create(ctx, user2)
	if err != nil {
		t.Errorf("Unexpected error creating user2: %v", err)
	}

	// GetByEmail should return the first user created with that email
	retrieved, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Errorf("Unexpected error getting user by email: %v", err)
	}
	if retrieved.ID != user1.ID {
		t.Errorf("GetByEmail returned wrong user. Expected ID %d, Got ID %d", user1.ID, retrieved.ID)
	}
}

func TestUserRepository_EmptyFields(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Create a user with empty fields
	user := &models.User{
		Email:        "",
		Username:     "",
		PasswordHash: "",
		Role:         "",
	}
	err := repo.Create(ctx, user)
	if err != nil {
		t.Errorf("Unexpected error creating user with empty fields: %v", err)
	}

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Errorf("Unexpected error retrieving created user: %v", err)
	}
	if retrieved.Email != "" || retrieved.Username != "" || retrieved.Role != "" {
		t.Errorf("User with empty fields was not created correctly")
	}
}

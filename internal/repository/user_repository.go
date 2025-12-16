package repository

import (
	"context"
	"database/sql"
	"errors"

	"golang-project/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	List(ctx context.Context, limit, offset int) ([]models.User, int, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role string) error
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, username, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, role, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
	).Scan(&user.ID, &user.Role, &user.CreatedAt, &user.UpdatedAt)
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var u models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	countQuery := "SELECT COUNT(*) FROM users"
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, username, password_hash, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *PostgresUserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET role = $1, updated_at = NOW()
		WHERE id = $2
	`, role, id)
	return err
}

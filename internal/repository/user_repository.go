package repository

import (
	"context"
	"database/sql"
	"errors"

	"golang-project/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
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

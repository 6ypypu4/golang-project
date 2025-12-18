package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang-project/internal/models"
)

type UserRepository interface {
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

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	var u models.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
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

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
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

func (r *PostgresUserRepository) List(ctx context.Context, filters models.UserFilters, limit, offset int) ([]models.User, int, error) {
	whereParts := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if filters.Search != "" {
		args = append(args, "%"+filters.Search+"%")
		whereParts = append(whereParts, fmt.Sprintf("(LOWER(email) LIKE LOWER($%d) OR LOWER(username) LIKE LOWER($%d))", argPos, argPos))
		argPos++
	}
	if filters.Role != "" {
		args = append(args, filters.Role)
		whereParts = append(whereParts, fmt.Sprintf("role = $%d", argPos))
		argPos++
	}

	whereSQL := strings.Join(whereParts, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	argsWithPage := append([]interface{}{}, args...)
	argsWithPage = append(argsWithPage, limit, offset)

	query := fmt.Sprintf(`
		SELECT id, email, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPos, argPos+1)

	rows, err := r.db.QueryContext(ctx, query, argsWithPage...)
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

func (r *PostgresUserRepository) UpdateRole(ctx context.Context, id int, role string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET role = $1, updated_at = NOW()
		WHERE id = $2
	`, role, id)
	return err
}

func (r *PostgresUserRepository) Update(ctx context.Context, id int, email, username string) error {
	query := `
		UPDATE users 
		SET email = COALESCE(NULLIF($1, ''), email),
		    username = COALESCE(NULLIF($2, ''), username),
		    updated_at = NOW()
		WHERE id = $3
	`
	result, err := r.db.ExecContext(ctx, query, email, username, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, id int, passwordHash string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`, passwordHash, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresUserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (r *PostgresUserRepository) CountLast7Days(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '7 days'").Scan(&count)
	return count, err
}

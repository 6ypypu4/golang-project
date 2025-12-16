package repository

import (
	"context"
	"database/sql"

	"golang-project/internal/models"

	"github.com/google/uuid"
)

type GenreRepository struct {
	db *sql.DB
}

func NewGenreRepository(db *sql.DB) *GenreRepository {
	return &GenreRepository{db: db}
}

func (r *GenreRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Genre, error) {
	var genre models.Genre
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, name, created_at FROM genres WHERE id = $1",
		id,
	).Scan(&genre.ID, &genre.Name, &genre.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &genre, nil
}

func (r *GenreRepository) GetByName(ctx context.Context, name string) (*models.Genre, error) {
	var genre models.Genre
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, name, created_at FROM genres WHERE name = $1",
		name,
	).Scan(&genre.ID, &genre.Name, &genre.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &genre, nil
}

func (r *GenreRepository) Create(ctx context.Context, genre *models.Genre) error {
	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO genres (name) VALUES ($1) RETURNING id, created_at",
		genre.Name,
	).Scan(&genre.ID, &genre.CreatedAt)
	return err
}

func (r *GenreRepository) Update(ctx context.Context, genre *models.Genre) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE genres SET name = $1 WHERE id = $2",
		genre.Name, genre.ID,
	)
	return err
}

func (r *GenreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM genres WHERE id = $1", id)
	return err
}

func (r *GenreRepository) GetAll(ctx context.Context) ([]models.Genre, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, name, created_at FROM genres ORDER BY name",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var genre models.Genre
		if err := rows.Scan(&genre.ID, &genre.Name, &genre.CreatedAt); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	return genres, rows.Err()
}

func (r *GenreRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM genres").Scan(&count)
	return count, err
}

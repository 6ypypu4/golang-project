package repository

import (
	"context"
	"database/sql"
	"fmt"

	"golang-project/internal/models"
)

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) GetByID(ctx context.Context, id int) (*models.Review, error) {
	var review models.Review
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, movie_id, user_id, rating, title, content, created_at, updated_at
		 FROM reviews WHERE id = $1`,
		id,
	).Scan(
		&review.ID, &review.MovieID, &review.UserID, &review.Rating,
		&review.Title, &review.Content, &review.CreatedAt, &review.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) GetByMovieID(ctx context.Context, movieID int, limit, offset int) ([]models.Review, error) {
	rows, err := r.db.QueryContext(
		ctx,
		fmt.Sprintf(
			`SELECT id, movie_id, user_id, rating, title, content, created_at, updated_at
			 FROM reviews
			 WHERE %s
			 ORDER BY created_at DESC
			 LIMIT $%d OFFSET $%d`,
			where, argPos, argPos+1,
		),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var review models.Review
		if err := rows.Scan(
			&review.ID, &review.MovieID, &review.UserID, &review.Rating,
			&review.Title, &review.Content, &review.CreatedAt, &review.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func (r *ReviewRepository) GetByUserID(ctx context.Context, userID int, limit, offset int) ([]models.Review, error) {
	rows, err := r.db.QueryContext(
		ctx,
		fmt.Sprintf(
			`SELECT id, movie_id, user_id, rating, title, content, created_at, updated_at
			 FROM reviews
			 WHERE %s
			 ORDER BY created_at DESC
			 LIMIT $%d OFFSET $%d`,
			where, argPos, argPos+1,
		),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var review models.Review
		if err := rows.Scan(
			&review.ID, &review.MovieID, &review.UserID, &review.Rating,
			&review.Title, &review.Content, &review.CreatedAt, &review.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func (r *ReviewRepository) GetByMovieAndUser(ctx context.Context, movieID, userID int) (*models.Review, error) {
	var review models.Review
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, movie_id, user_id, rating, title, content, created_at, updated_at
		 FROM reviews WHERE movie_id = $1 AND user_id = $2`,
		movieID, userID,
	).Scan(
		&review.ID, &review.MovieID, &review.UserID, &review.Rating,
		&review.Title, &review.Content, &review.CreatedAt, &review.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) Create(ctx context.Context, review *models.Review) error {
	return r.db.QueryRowContext(
		ctx,
		`INSERT INTO reviews (movie_id, user_id, rating, title, content)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		review.MovieID, review.UserID, review.Rating, review.Title, review.Content,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
}

func (r *ReviewRepository) Update(ctx context.Context, review *models.Review) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE reviews 
		 SET rating = $1, title = $2, content = $3, updated_at = NOW()
		 WHERE id = $4`,
		review.Rating, review.Title, review.Content, review.ID,
	)
	return err
}

func (r *ReviewRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM reviews WHERE id = $1", id)
	return err
}

func (r *ReviewRepository) CountByMovieID(ctx context.Context, movieID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM reviews WHERE movie_id = $1",
		movieID,
	).Scan(&count)
	return count, err
}

func (r *ReviewRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM reviews WHERE user_id = $1",
		userID,
	).Scan(&count)
	return count, err
}

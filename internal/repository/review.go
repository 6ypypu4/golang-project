package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

func (r *ReviewRepository) GetByMovieID(ctx context.Context, movieID int, filters models.ReviewFilters, limit, offset int) ([]models.Review, error) {
	whereParts := []string{"movie_id = $1"}
	args := []interface{}{movieID}
	argPos := 2

	if filters.MinRating > 0 {
		args = append(args, filters.MinRating)
		whereParts = append(whereParts, fmt.Sprintf("rating >= $%d", argPos))
		argPos++
	}
	if filters.MaxRating > 0 {
		args = append(args, filters.MaxRating)
		whereParts = append(whereParts, fmt.Sprintf("rating <= $%d", argPos))
		argPos++
	}

	whereSQL := strings.Join(whereParts, " AND ")

	orderBy := "created_at DESC"
	switch filters.Sort {
	case "rating_desc":
		orderBy = "rating DESC"
	case "rating_asc":
		orderBy = "rating ASC"
	case "created_desc":
		orderBy = "created_at DESC"
	case "created_asc":
		orderBy = "created_at ASC"
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id, movie_id, user_id, rating, title, content, created_at, updated_at
		FROM reviews
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, orderBy, argPos, argPos+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *ReviewRepository) GetByUserID(ctx context.Context, userID int, filters models.ReviewFilters, limit, offset int) ([]models.Review, error) {
	whereParts := []string{"user_id = $1"}
	args := []interface{}{userID}
	argPos := 2

	if filters.MinRating > 0 {
		args = append(args, filters.MinRating)
		whereParts = append(whereParts, fmt.Sprintf("rating >= $%d", argPos))
		argPos++
	}
	if filters.MaxRating > 0 {
		args = append(args, filters.MaxRating)
		whereParts = append(whereParts, fmt.Sprintf("rating <= $%d", argPos))
		argPos++
	}

	whereSQL := strings.Join(whereParts, " AND ")

	orderBy := "created_at DESC"
	switch filters.Sort {
	case "rating_desc":
		orderBy = "rating DESC"
	case "rating_asc":
		orderBy = "rating ASC"
	case "created_desc":
		orderBy = "created_at DESC"
	case "created_asc":
		orderBy = "created_at ASC"
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id, movie_id, user_id, rating, title, content, created_at, updated_at
		FROM reviews
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, orderBy, argPos, argPos+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *ReviewRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reviews").Scan(&count)
	return count, err
}

func (r *ReviewRepository) CountLast7Days(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reviews WHERE created_at >= NOW() - INTERVAL '7 days'").Scan(&count)
	return count, err
}

func (r *ReviewRepository) GetAverageRatingByUserID(ctx context.Context, userID int) (float64, error) {
	var avg sql.NullFloat64
	err := r.db.QueryRowContext(
		ctx,
		"SELECT AVG(rating) FROM reviews WHERE user_id = $1",
		userID,
	).Scan(&avg)
	if err != nil {
		return 0, err
	}
	if !avg.Valid {
		return 0, nil
	}
	return avg.Float64, nil
}

func (r *ReviewRepository) GetFavoriteGenreByUserID(ctx context.Context, userID int) (*models.Genre, error) {
	var genre models.Genre
	err := r.db.QueryRowContext(
		ctx,
		`SELECT g.id, g.name, g.created_at
		 FROM genres g
		 INNER JOIN movie_genres mg ON g.id = mg.genre_id
		 INNER JOIN reviews r ON r.movie_id = mg.movie_id
		 WHERE r.user_id = $1
		 GROUP BY g.id, g.name, g.created_at
		 ORDER BY COUNT(*) DESC
		 LIMIT 1`,
		userID,
	).Scan(&genre.ID, &genre.Name, &genre.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &genre, nil
}

package repository

import (
	"database/sql"

	"golang-project/internal/models"

	"github.com/google/uuid"
)

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) GetByID(id uuid.UUID) (*models.Review, error) {
	var review models.Review
	err := r.db.QueryRow(
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

func (r *ReviewRepository) GetByMovieID(movieID uuid.UUID, limit, offset int) ([]models.Review, error) {
	rows, err := r.db.Query(
		`SELECT id, movie_id, user_id, rating, title, content, created_at, updated_at
		 FROM reviews
		 WHERE movie_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		movieID, limit, offset,
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

func (r *ReviewRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Review, error) {
	rows, err := r.db.Query(
		`SELECT id, movie_id, user_id, rating, title, content, created_at, updated_at
		 FROM reviews
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
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

func (r *ReviewRepository) GetByMovieAndUser(movieID, userID uuid.UUID) (*models.Review, error) {
	var review models.Review
	err := r.db.QueryRow(
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

func (r *ReviewRepository) Create(review *models.Review) error {
	err := r.db.QueryRow(
		`INSERT INTO reviews (movie_id, user_id, rating, title, content)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		review.MovieID, review.UserID, review.Rating, review.Title, review.Content,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
	return err
}

func (r *ReviewRepository) Update(review *models.Review) error {
	_, err := r.db.Exec(
		`UPDATE reviews 
		 SET rating = $1, title = $2, content = $3, updated_at = NOW()
		 WHERE id = $4`,
		review.Rating, review.Title, review.Content, review.ID,
	)
	return err
}

func (r *ReviewRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM reviews WHERE id = $1", id)
	return err
}

func (r *ReviewRepository) CountByMovieID(movieID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM reviews WHERE movie_id = $1",
		movieID,
	).Scan(&count)
	return count, err
}

func (r *ReviewRepository) CountByUserID(userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM reviews WHERE user_id = $1",
		userID,
	).Scan(&count)
	return count, err
}

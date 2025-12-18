package repository

import (
	"context"
	"database/sql"

	"golang-project/internal/models"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Insert(ctx context.Context, log *models.AuditLog) error {
	var userID interface{}
	var movieID interface{}
	var reviewID interface{}

	if log.UserID != nil {
		userID = *log.UserID
	}
	if log.MovieID != nil {
		movieID = *log.MovieID
	}
	if log.ReviewID != nil {
		reviewID = *log.ReviewID
	}

	return r.db.QueryRowContext(
		ctx,
		`INSERT INTO audit_logs (user_id, movie_id, review_id, event, details)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		userID, movieID, reviewID, log.Event, log.Details,
	).Scan(&log.ID, &log.CreatedAt)
}



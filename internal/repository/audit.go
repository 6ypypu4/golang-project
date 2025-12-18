package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func (r *AuditRepository) List(ctx context.Context, filters models.AuditLogFilters, limit, offset int) ([]models.AuditLog, int, error) {
	whereParts := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if filters.Event != "" {
		args = append(args, filters.Event)
		whereParts = append(whereParts, "event = $"+fmt.Sprintf("%d", argPos))
		argPos++
	}
	if filters.UserID != nil {
		args = append(args, *filters.UserID)
		whereParts = append(whereParts, "user_id = $"+fmt.Sprintf("%d", argPos))
		argPos++
	}
	if filters.FromDate != nil {
		args = append(args, *filters.FromDate)
		whereParts = append(whereParts, "created_at >= $"+fmt.Sprintf("%d", argPos))
		argPos++
	}
	if filters.ToDate != nil {
		args = append(args, *filters.ToDate)
		whereParts = append(whereParts, "created_at <= $"+fmt.Sprintf("%d", argPos))
		argPos++
	}

	whereSQL := strings.Join(whereParts, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	argsWithPage := append([]interface{}{}, args...)
	argsWithPage = append(argsWithPage, limit, offset)

	query := fmt.Sprintf(`
		SELECT id, user_id, movie_id, review_id, event, details, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPos, argPos+1)

	rows, err := r.db.QueryContext(ctx, query, argsWithPage...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var userID, movieID, reviewID sql.NullInt64
		if err := rows.Scan(
			&log.ID, &userID, &movieID, &reviewID,
			&log.Event, &log.Details, &log.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if userID.Valid {
			uid := int(userID.Int64)
			log.UserID = &uid
		}
		if movieID.Valid {
			mid := int(movieID.Int64)
			log.MovieID = &mid
		}
		if reviewID.Valid {
			rid := int(reviewID.Int64)
			log.ReviewID = &rid
		}
		logs = append(logs, log)
	}
	return logs, total, rows.Err()
}

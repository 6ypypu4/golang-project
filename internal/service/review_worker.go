package service

import (
	"context"
	"log"
	"time"

	"golang-project/internal/models"
)

type ReviewEventType string

const (
	EventReviewCreated ReviewEventType = "review_created"
	EventReviewUpdated ReviewEventType = "review_updated"
	EventReviewDeleted ReviewEventType = "review_deleted"
)

type ReviewEvent struct {
	Type     ReviewEventType
	MovieID  int
	UserID   int
	ReviewID int
	Time     time.Time
}

type MovieRater interface {
	UpdateAverageRating(ctx context.Context, movieID int) error
}

type AuditWriter interface {
	Insert(ctx context.Context, log *models.AuditLog) error
}

func StartReviewWorker(ctx context.Context, events <-chan ReviewEvent, movies MovieRater, audit AuditWriter) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-events:
				handleReviewEvent(ctx, e, movies, audit)
			}
		}
	}()
}

func handleReviewEvent(ctx context.Context, e ReviewEvent, movies MovieRater, audit AuditWriter) {
	if e.MovieID != 0 {
		if err := movies.UpdateAverageRating(ctx, e.MovieID); err != nil {
			log.Printf("review worker: update average rating error: %v", err)
		}
	}

	if audit == nil {
		return
	}

	var userID *int
	var movieID *int
	var reviewID *int

	if e.UserID != 0 {
		id := e.UserID
		userID = &id
	}
	if e.MovieID != 0 {
		id := e.MovieID
		movieID = &id
	}
	if e.ReviewID != 0 {
		id := e.ReviewID
		reviewID = &id
	}

	logEntry := &models.AuditLog{
		UserID:   userID,
		MovieID:  movieID,
		ReviewID: reviewID,
		Event:    string(e.Type),
		Details:  "",
	}

	if err := audit.Insert(ctx, logEntry); err != nil {
		log.Printf("review worker: audit insert error: %v", err)
	}
}



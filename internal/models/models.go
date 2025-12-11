package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Genre struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Movie struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description" db:"description"`
	ReleaseYear     int       `json:"release_year" db:"release_year"`
	Director        string    `json:"director" db:"director"`
	DurationMinutes int       `json:"duration_minutes" db:"duration_minutes"`
	AverageRating   float64   `json:"average_rating" db:"average_rating"`
	Genres          []Genre   `json:"genres,omitempty"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type MovieGenre struct {
	MovieID uuid.UUID `json:"movie_id" db:"movie_id"`
	GenreID uuid.UUID `json:"genre_id" db:"genre_id"`
}

type Review struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MovieID   uuid.UUID `json:"movie_id" db:"movie_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	User      *User     `json:"user,omitempty"`
	Movie     *Movie    `json:"movie,omitempty"`
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateMovieRequest struct {
	Title           string   `json:"title" validate:"required,max=255"`
	Description     string   `json:"description"`
	ReleaseYear     int      `json:"release_year" validate:"required,min=1800,max=2030"`
	Director        string   `json:"director" validate:"max=255"`
	DurationMinutes int      `json:"duration_minutes" validate:"min=1"`
	GenreIDs        []string `json:"genre_ids" validate:"required,min=1"`
}

type UpdateMovieRequest struct {
	Title           string   `json:"title" validate:"max=255"`
	Description     string   `json:"description"`
	ReleaseYear     int      `json:"release_year" validate:"min=1800,max=2030"`
	Director        string   `json:"director" validate:"max=255"`
	DurationMinutes int      `json:"duration_minutes" validate:"min=1"`
	GenreIDs        []string `json:"genre_ids"`
}

type CreateGenreRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=10"`
	Title   string `json:"title" validate:"required,max=255"`
	Content string `json:"content" validate:"required"`
}

type UpdateReviewRequest struct {
	Rating  int    `json:"rating" validate:"min=1,max=10"`
	Title   string `json:"title" validate:"max=255"`
	Content string `json:"content"`
}

type PaginationParams struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

type MovieFilters struct {
	Genre     string     `json:"genre"`
	GenreID   *uuid.UUID `json:"-"`
	Year      int        `json:"year"`
	MinRating float64    `json:"min_rating"`
	Search    string     `json:"search"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

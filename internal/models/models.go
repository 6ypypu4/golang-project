package models

import (
	"time"
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Genre struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Movie struct {
	ID              int       `json:"id" db:"id"`
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
	MovieID int `json:"movie_id" db:"movie_id"`
	GenreID int `json:"genre_id" db:"genre_id"`
}

type Review struct {
	ID        int       `json:"id" db:"id"`
	MovieID   int       `json:"movie_id" db:"movie_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	User      *User     `json:"user,omitempty"`
	Movie     *Movie    `json:"movie,omitempty"`
}

type AuditLog struct {
	ID        int       `json:"id" db:"id"`
	UserID    *int      `json:"user_id" db:"user_id"`
	MovieID   *int      `json:"movie_id" db:"movie_id"`
	ReviewID  *int      `json:"review_id" db:"review_id"`
	Event     string    `json:"event" db:"event"`
	Details   string    `json:"details" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Username string `json:"username" validate:"omitempty,min=3,max=100"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
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
	Genre     string  `json:"genre"`
	GenreID   *int    `json:"-"`
	Year      int     `json:"year"`
	MinRating float64 `json:"min_rating"`
	Search    string  `json:"search"`
	Sort      string  `json:"sort"`
}

type ReviewFilters struct {
	MinRating int    `json:"min_rating"`
	MaxRating int    `json:"max_rating"`
	Sort      string `json:"sort"`
}

type UserFilters struct {
	Search string `json:"search"`
	Role   string `json:"role"`
}

type AuditLogFilters struct {
	Event    string     `json:"event"`
	UserID   *int       `json:"user_id"`
	FromDate *time.Time `json:"from_date"`
	ToDate   *time.Time `json:"to_date"`
}

type UserStats struct {
	AverageRating float64 `json:"average_rating"`
	FavoriteGenre *Genre  `json:"favorite_genre,omitempty"`
}

type AdminStats struct {
	TotalUsers       int     `json:"total_users"`
	TotalMovies      int     `json:"total_movies"`
	TotalReviews     int     `json:"total_reviews"`
	TotalGenres      int     `json:"total_genres"`
	AverageRating    float64 `json:"average_rating"`
	UsersLast7Days   int     `json:"users_last_7_days"`
	ReviewsLast7Days int     `json:"reviews_last_7_days"`
	MoviesLast7Days  int     `json:"movies_last_7_days"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

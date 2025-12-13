package repository

import (
	"database/sql"

	"golang-project/internal/models"

	"github.com/google/uuid"
)

type MovieRepository struct {
	db *sql.DB
}

func NewMovieRepository(db *sql.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

func (r *MovieRepository) GetByID(id uuid.UUID) (*models.Movie, error) {
	var movie models.Movie
	err := r.db.QueryRow(
		`SELECT id, title, description, release_year, director, duration_minutes, 
		 average_rating, created_at, updated_at 
		 FROM movies WHERE id = $1`,
		id,
	).Scan(
		&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseYear,
		&movie.Director, &movie.DurationMinutes, &movie.AverageRating,
		&movie.CreatedAt, &movie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func (r *MovieRepository) Create(movie *models.Movie) error {
	err := r.db.QueryRow(
		`INSERT INTO movies (title, description, release_year, director, duration_minutes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		movie.Title, movie.Description, movie.ReleaseYear,
		movie.Director, movie.DurationMinutes,
	).Scan(&movie.ID, &movie.CreatedAt, &movie.UpdatedAt)
	return err
}

func (r *MovieRepository) Update(movie *models.Movie) error {
	_, err := r.db.Exec(
		`UPDATE movies 
		 SET title = $1, description = $2, release_year = $3, 
		     director = $4, duration_minutes = $5, updated_at = NOW()
		 WHERE id = $6`,
		movie.Title, movie.Description, movie.ReleaseYear,
		movie.Director, movie.DurationMinutes, movie.ID,
	)
	return err
}

func (r *MovieRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM movies WHERE id = $1", id)
	return err
}

func (r *MovieRepository) GetAll(limit, offset int) ([]models.Movie, error) {
	rows, err := r.db.Query(
		`SELECT id, title, description, release_year, director, duration_minutes,
		 average_rating, created_at, updated_at
		 FROM movies
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		if err := rows.Scan(
			&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseYear,
			&movie.Director, &movie.DurationMinutes, &movie.AverageRating,
			&movie.CreatedAt, &movie.UpdatedAt,
		); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, rows.Err()
}

func (r *MovieRepository) GetGenresByMovieID(movieID uuid.UUID) ([]models.Genre, error) {
	rows, err := r.db.Query(
		`SELECT g.id, g.name, g.created_at
		 FROM genres g
		 INNER JOIN movie_genres mg ON g.id = mg.genre_id
		 WHERE mg.movie_id = $1`,
		movieID,
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

func (r *MovieRepository) AddGenre(movieID, genreID uuid.UUID) error {
	_, err := r.db.Exec(
		"INSERT INTO movie_genres (movie_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		movieID, genreID,
	)
	return err
}

func (r *MovieRepository) RemoveGenre(movieID, genreID uuid.UUID) error {
	_, err := r.db.Exec(
		"DELETE FROM movie_genres WHERE movie_id = $1 AND genre_id = $2",
		movieID, genreID,
	)
	return err
}

func (r *MovieRepository) SetGenres(movieID uuid.UUID, genreIDs []uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Удаляем все существующие связи
	_, err = tx.Exec("DELETE FROM movie_genres WHERE movie_id = $1", movieID)
	if err != nil {
		return err
	}

	// Добавляем новые связи
	for _, genreID := range genreIDs {
		_, err = tx.Exec(
			"INSERT INTO movie_genres (movie_id, genre_id) VALUES ($1, $2)",
			movieID, genreID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MovieRepository) UpdateAverageRating(movieID uuid.UUID) error {
	_, err := r.db.Exec(
		`UPDATE movies 
		 SET average_rating = (
			 SELECT COALESCE(AVG(rating), 0)
			 FROM reviews
			 WHERE movie_id = $1
		 )
		 WHERE id = $1`,
		movieID,
	)
	return err
}

func (r *MovieRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM movies").Scan(&count)
	return count, err
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"golang-project/internal/models"
)

type MovieRepository struct {
	db *sql.DB
}

func NewMovieRepository(db *sql.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

func (r *MovieRepository) GetByID(ctx context.Context, id int) (*models.Movie, error) {
	var movie models.Movie
	err := r.db.QueryRowContext(
		ctx,
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

func (r *MovieRepository) Create(ctx context.Context, movie *models.Movie) error {
	return r.db.QueryRowContext(
		ctx,
		`INSERT INTO movies (title, description, release_year, director, duration_minutes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		movie.Title, movie.Description, movie.ReleaseYear,
		movie.Director, movie.DurationMinutes,
	).Scan(&movie.ID, &movie.CreatedAt, &movie.UpdatedAt)
}

func (r *MovieRepository) Update(ctx context.Context, movie *models.Movie) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE movies 
		 SET title = $1, description = $2, release_year = $3, 
		     director = $4, duration_minutes = $5, updated_at = NOW()
		 WHERE id = $6`,
		movie.Title, movie.Description, movie.ReleaseYear,
		movie.Director, movie.DurationMinutes, movie.ID,
	)
	return err
}

func (r *MovieRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM movies WHERE id = $1", id)
	return err
}

func (r *MovieRepository) List(ctx context.Context, filters models.MovieFilters, limit, offset int) ([]models.Movie, int, error) {
	whereParts := []string{"1=1"}
	args := []interface{}{}

	if filters.GenreID != nil {
		args = append(args, *filters.GenreID)
		whereParts = append(whereParts, fmt.Sprintf("EXISTS (SELECT 1 FROM movie_genres mg WHERE mg.movie_id = m.id AND mg.genre_id = $%d)", len(args)))
	}
	if filters.Genre != "" {
		args = append(args, "%"+filters.Genre+"%")
		whereParts = append(whereParts, fmt.Sprintf("EXISTS (SELECT 1 FROM genres g INNER JOIN movie_genres mg ON g.id = mg.genre_id WHERE mg.movie_id = m.id AND LOWER(g.name) LIKE LOWER($%d))", len(args)))
	}
	if filters.Year != 0 {
		args = append(args, filters.Year)
		whereParts = append(whereParts, fmt.Sprintf("m.release_year = $%d", len(args)))
	}
	if filters.MinRating > 0 {
		args = append(args, filters.MinRating)
		whereParts = append(whereParts, fmt.Sprintf("m.average_rating >= $%d", len(args)))
	}
	if filters.Search != "" {
		args = append(args, "%"+filters.Search+"%")
		whereParts = append(whereParts, fmt.Sprintf("(LOWER(m.title) LIKE LOWER($%d) OR LOWER(m.description) LIKE LOWER($%d))", len(args), len(args)))
	}

	whereSQL := strings.Join(whereParts, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(DISTINCT m.id) FROM movies m WHERE %s`, whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	argsWithPage := append([]interface{}{}, args...)
	argsWithPage = append(argsWithPage, limit, offset)

	query := fmt.Sprintf(
		`SELECT DISTINCT m.id, m.title, m.description, m.release_year, m.director, m.duration_minutes,
		        m.average_rating, m.created_at, m.updated_at
		 FROM movies m
		 WHERE %s
		 ORDER BY m.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereSQL, len(args)+1, len(args)+2,
	)
	rows, err := r.db.QueryContext(ctx, query, argsWithPage...)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}
		movies = append(movies, movie)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	for i := range movies {
		genres, err := r.GetGenresByMovieID(ctx, movies[i].ID)
		if err != nil {
			return nil, 0, err
		}
		movies[i].Genres = genres
	}

	return movies, total, nil
}

func (r *MovieRepository) GetGenresByMovieID(ctx context.Context, movieID int) ([]models.Genre, error) {
	rows, err := r.db.QueryContext(
		ctx,
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

func (r *MovieRepository) SetGenres(ctx context.Context, movieID int, genreIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM movie_genres WHERE movie_id = $1", movieID); err != nil {
		return err
	}

	for _, genreID := range genreIDs {
		if _, err := tx.ExecContext(ctx, "INSERT INTO movie_genres (movie_id, genre_id) VALUES ($1, $2)", movieID, genreID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MovieRepository) UpdateAverageRating(ctx context.Context, movieID int) error {
	_, err := r.db.ExecContext(
		ctx,
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

func (r *MovieRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM movies").Scan(&count)
	return count, err
}

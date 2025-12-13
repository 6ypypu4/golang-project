package repository

import (
	"database/sql"
)

// Repositories хранит все репозитории
type Repositories struct {
	User   *UserRepository
	Movie  *MovieRepository
	Genre  *GenreRepository
	Review *ReviewRepository
}

// NewRepositories создаёт и возвращает все репозитории
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:   NewUserRepository(db),
		Movie:  NewMovieRepository(db),
		Genre:  NewGenreRepository(db),
		Review: NewReviewRepository(db),
	}
}

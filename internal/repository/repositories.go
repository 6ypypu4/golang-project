package repository

import (
	"database/sql"
)

// Repositories С:С?Р°Р?РёС' Р?С?Рч С?РчРїР?Р·РёС'Р?С?РёРё
type Repositories struct {
	User   UserRepository
	Movie  *MovieRepository
	Genre  *GenreRepository
	Review *ReviewRepository
	Audit  *AuditRepository
}

// NewRepositories С?Р?Р·Р?Р°С'С' Рё Р?Р?Р·Р?С?Р°С%Р°РчС' Р?С?Рч С?РчРїР?Р·РёС'Р?С?РёРё
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:   NewUserRepository(db),
		Movie:  NewMovieRepository(db),
		Genre:  NewGenreRepository(db),
		Review: NewReviewRepository(db),
		Audit:  NewAuditRepository(db),
	}
}

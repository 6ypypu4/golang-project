package database

import (
	"database/sql"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	// TODO: Initialize database connection
	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}


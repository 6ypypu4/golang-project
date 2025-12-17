package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(ctx context.Context, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping db: %w", err)
	}

	DB = db
	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func RunMigrations(path string) error {
	if DB == nil {
		return fmt.Errorf("database is not initialized")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.ToSlash(absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("new migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

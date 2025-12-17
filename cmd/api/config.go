package main

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBDsn          string
	JWTSecret      string
	MigrationsPath string
}

type ErrMissingEnv string

func (e ErrMissingEnv) Error() string {
	return "missing environment variable: " + string(e)
}

func loadConfig() (*Config, error) {
	// Load .env file if it exists (silently ignore if not found)
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, ErrMissingEnv("DB_DSN")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, ErrMissingEnv("JWT_SECRET")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "internal/migrations"
	}

	return &Config{
		Port:           port,
		DBDsn:          dsn,
		JWTSecret:      secret,
		MigrationsPath: migrationsPath,
	}, nil
}

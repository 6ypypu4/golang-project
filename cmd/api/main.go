package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-project/internal/database"
	"golang-project/internal/handler"
	"golang-project/internal/repository"
)

type Config struct {
	Port      string
	DBDsn     string
	JWTSecret string
}

func loadConfig() (*Config, error) {
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

	return &Config{
		Port:      port,
		DBDsn:     dsn,
		JWTSecret: secret,
	}, nil
}

type ErrMissingEnv string

func (e ErrMissingEnv) Error() string {
	return "missing environment variable: " + string(e)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := database.InitDB(dbCtx, cfg.DBDsn); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer func() {
		if err := database.CloseDB(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	// Создаём репозитории
	repos := repository.NewRepositories(database.DB)

	// Настраиваем роуты с репозиториями
	router := handler.SetupRoutes(repos)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("API server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server exited gracefully")
}

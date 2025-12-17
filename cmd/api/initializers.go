package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang-project/internal/database"
	"golang-project/internal/handler"
)

// AppInitializer initializes application components
type AppInitializer struct {
	config *Config
	db     *sql.DB
	router http.Handler
	server *http.Server
}

func NewAppInitializer() *AppInitializer {
	return &AppInitializer{}
}

// InitializeConfig loads application configuration
func (ai *AppInitializer) InitializeConfig() error {
	log.Println("loading configuration")

	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ai.config = cfg
	log.Printf("configuration loaded (port=%s)", cfg.Port)
	return nil
}

// InitializeDatabase initializes database connection and runs migrations
func (ai *AppInitializer) InitializeDatabase(ctx context.Context) error {
	log.Println("initializing database")

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := database.InitDB(dbCtx, ai.config.DBDsn); err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	ai.db = database.DB
	log.Println("database connected")

	if err := database.RunMigrations(ai.config.MigrationsPath); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Println("migrations completed")
	return nil
}

// InitializeRouter sets up HTTP routes
func (ai *AppInitializer) InitializeRouter() error {
	if ai.db == nil {
		return fmt.Errorf("database not initialized")
	}

	log.Println("initializing router")
	ai.router = handler.SetupRoutes(ai.db, ai.config.JWTSecret)
	return nil
}

// Initialize HTTP server
func (ai *AppInitializer) InitializeServer() error {
	if ai.router == nil {
		return fmt.Errorf("router not initialized")
	}

	ai.server = &http.Server{
		Addr:         ":" + ai.config.Port,
		Handler:      ai.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("server configured (port=%s)", ai.config.Port)
	return nil
}

// Starts HTTP server
func (ai *AppInitializer) StartServer() error {
	if ai.server == nil {
		return fmt.Errorf("server not initialized")
	}

	log.Printf("starting server on :%s", ai.config.Port)

	go func() {
		if err := ai.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	return nil
}

// Shutdown server and database
func (ai *AppInitializer) Shutdown(ctx context.Context) error {
	if ai.server != nil {
		if err := ai.server.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}

	if err := database.CloseDB(); err != nil {
		log.Printf("database close error: %v", err)
	}

	return nil
}

// Getters
func (ai *AppInitializer) GetConfig() *Config {
	return ai.config
}

func (ai *AppInitializer) GetDB() *sql.DB {
	return ai.db
}

func (ai *AppInitializer) GetServer() *http.Server {
	return ai.server
}

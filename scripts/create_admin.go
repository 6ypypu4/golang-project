package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run create_admin.go <email> <password> [username]")
		fmt.Println("Example: go run create_admin.go admin@example.com admin123")
		fmt.Println("Example: go run create_admin.go admin@example.com admin123 myadmin")
		os.Exit(1)
	}

	if len(os.Args) > 4 {
		fmt.Println("Error: Too many arguments")
		fmt.Println("Usage: go run create_admin.go <email> <password> [username]")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	username := "admin"
	if len(os.Args) == 4 {
		username = os.Args[3]
	}

	// Валидация email
	if !strings.Contains(email, "@") {
		fmt.Printf("Error: Invalid email format: %s\n", email)
		fmt.Println("Email must contain @ symbol")
		os.Exit(1)
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://app:app@localhost:5432/app?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	query := `
		INSERT INTO users (email, username, password_hash, role)
		VALUES ($1, $2, $3, 'admin')
		ON CONFLICT (email) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
		    role = 'admin',
		    updated_at = NOW()
		RETURNING id, email, username, role
	`

	var userID int
	var userEmail, userUsername, userRole string
	err = db.QueryRowContext(ctx, query, email, username, string(hash)).Scan(
		&userID, &userEmail, &userUsername, &userRole,
	)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Admin user created successfully!\n")
	fmt.Printf("ID: %d\n", userID)
	fmt.Printf("Email: %s\n", userEmail)
	fmt.Printf("Username: %s\n", userUsername)
	fmt.Printf("Role: %s\n", userRole)
}

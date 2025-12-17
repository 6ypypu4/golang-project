package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	initializer := NewAppInitializer()

	log.Println("application starting")
	//init each of the component
	if err := initializer.InitializeConfig(); err != nil {
		log.Fatalf("init config: %v", err)
	}

	if err := initializer.InitializeDatabase(ctx); err != nil {
		log.Fatalf("init database: %v", err)
	}

	if err := initializer.InitializeRouter(); err != nil {
		log.Fatalf("init router: %v", err)
	}

	if err := initializer.InitializeServer(); err != nil {
		log.Fatalf("init server: %v", err)
	}

	if err := initializer.StartServer(); err != nil {
		log.Fatalf("start server: %v", err)
	}

	log.Println("application started")

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := initializer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	log.Println("application stopped")
}

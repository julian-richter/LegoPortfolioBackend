package main

import (
	"LegoManagerAPI/internal/config"
	"LegoManagerAPI/internal/database"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
)

func main() {
	// Load configuration
	log.Info("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}
	log.Info("Configuration loaded successfully")

	// Initialize database connection
	log.Info("Connecting to database...")
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	log.Info("Database connection established")

	// Ping database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		log.Fatal("Failed to ping database", "error", err)
	}
	log.Info("Database ping successful!")

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Application is running. Press Ctrl+C to exit.")
	<-quit

	log.Info("Shutting down gracefully...")

	// Close database connection
	if err := db.Close(); err != nil {
		log.Error("Error closing database", "error", err)
	}

	log.Info("Shutdown complete")
}

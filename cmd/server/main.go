package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"LegoManagerAPI/internal/api"
	"LegoManagerAPI/internal/api/service"
	"LegoManagerAPI/internal/cache"
	"LegoManagerAPI/internal/config"
	"LegoManagerAPI/internal/config/application"
	"LegoManagerAPI/internal/database"

	"github.com/charmbracelet/log"
)

func init() {
	// Try to load the .env filr from project root
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found")
	}
}

func main() {
	// Load configuration
	log.Info("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}
	application.SetupLogger(cfg.App.LogLVL)
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

	// Initialize Redis connection
	log.Info("Connecting to Redis...")
	redisClient, err := cache.NewRedisClient(cfg.Cache)
	if err != nil {
		log.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redisClient.Close()

	// Initialize Bricklink service
	bricklinkService := service.NewBricklinkService(cfg.Bricklink)
	log.Info("Bricklink service initialized")

	// Create HTTP server
	server := api.NewServer(cfg, db, redisClient, bricklinkService)

	// Start the server in goroutine so it doesn't block
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	log.Info("Application is running. Press Ctrl+C to exit.")

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Server shutdown error", "error", err)
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Error("Error closing Redis", "error", err)
	}

	// Close database connection
	if err := db.Close(); err != nil {
		log.Error("Error closing database", "error", err)
	}

	log.Info("Shutdown complete")
}

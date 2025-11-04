package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"LegoManagerAPI/internal/api"
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
	Cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}
	application.SetupLogger(Cfg.App.LogLVL)
	log.Info("Configuration loaded successfully")

	// Initialize database connection
	log.Info("Connecting to database...")
	db, err := database.NewPostgresDB(Cfg.Database)
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
	redisClient, err := cache.NewRedisClient(Cfg.Cache)
	if err != nil {
		log.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redisClient.Close()

	server := api.NewServer(Cfg, db, redisClient)

	// start the server in goroutine for performance best practice and so it doesnt block requests
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

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

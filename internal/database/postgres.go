package database

import (
	"context"
	"fmt"
	"time"

	"LegoManagerAPI/internal/config/database"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

// NewPostgresDB initializes and returns a PostgresDB instance with a connection pool configured using the provided DatabaseConfig.
func NewPostgresDB(cfg database.DatabaseConfig) (*PostgresDB, error) {
	// Build the connection string
	connectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.MaxConns, cfg.MinConns)

	log.Debug("Attempting to connect to database", "connection_string", connectionString)

	// Parse Config
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure connection the pool
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Minute * 5

	// Create the connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	log.Info("Database connection pool created")

	return &PostgresDB{Pool: pool}, nil
}

// Ping checks the connection to the database by pinging the connection pool. Returns an error if the ping fails.
func (db *PostgresDB) Ping(ctx context.Context) error {
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// Close gracefully closes all active connections in the pool.
func (db *PostgresDB) Close() error {
	log.Info("Closing database connection pool")
	db.Pool.Close()
	log.Info("Database connection pool closed")
	return nil
}

// Stats return real-time statistics of the connection pool for monitoring purposes.
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

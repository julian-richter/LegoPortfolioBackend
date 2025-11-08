package database_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"LegoManagerAPI/internal/config/database"
	dbpkg "LegoManagerAPI/internal/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfig reads database credentials from environment variables for local testing.
func setupTestConfig() database.DatabaseConfig {
	port := 5432
	if p := os.Getenv("POSTGRES_PORT"); p != "" {
		// ignore conversion error for brevity
		fmt.Sscanf(p, "%d", &port)
	}

	return database.DatabaseConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     port,
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  "disable",
		MaxConns: 5,
		MinConns: 1,
	}
}

func TestNewPostgresDB_Success(t *testing.T) {
	cfg := setupTestConfig()

	db, err := dbpkg.NewPostgresDB(cfg)
	require.NoError(t, err, "should create Postgres connection pool")
	require.NotNil(t, db.Pool, "pool should not be nil")

	defer db.Close()
}

func TestPostgresDB_Ping(t *testing.T) {
	cfg := setupTestConfig()
	db, err := dbpkg.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	assert.NoError(t, err, "ping should succeed on healthy DB")
}

func TestPostgresDB_Stats(t *testing.T) {
	cfg := setupTestConfig()
	db, err := dbpkg.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	stats := db.Stats()
	assert.NotNil(t, stats, "stats should not be nil")
	assert.GreaterOrEqual(t, stats.TotalConns(), int32(0))
}

func TestPostgresDB_Close(t *testing.T) {
	cfg := setupTestConfig()
	db, err := dbpkg.NewPostgresDB(cfg)
	require.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err, "close should not return error")
}

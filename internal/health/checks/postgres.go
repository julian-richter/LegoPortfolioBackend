package checks

import (
	"context"
	"time"

	"LegoManagerAPI/internal/database"
	"LegoManagerAPI/internal/health"
)

type PostgresCheck struct {
	db *database.PostgresDB
}

func NewPostgresCheck(db *database.PostgresDB) *PostgresCheck {
	return &PostgresCheck{db: db}
}

func (p *PostgresCheck) Name() string {
	return "postgres"
}

func (p *PostgresCheck) Check(ctx context.Context) health.Status {
	start := time.Now()

	if err := p.db.Ping(ctx); err != nil {
		return health.Status{
			Status: "unhealthy",
			Error:  err.Error(),
		}
	}

	return health.Status{
		Status:  "healthy",
		Latency: time.Since(start).String(),
	}
}

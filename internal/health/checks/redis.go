package checks

import (
	"context"
	"time"

	"LegoManagerAPI/internal/cache"
	"LegoManagerAPI/internal/health"
)

type RedisCheck struct {
	client *cache.RedisClient
}

func NewRedisCheck(client *cache.RedisClient) *RedisCheck {
	return &RedisCheck{client: client}
}

func (r *RedisCheck) Name() string {
	return "redis"
}

func (r *RedisCheck) Check(ctx context.Context) health.Status {
	start := time.Now()

	if err := r.client.Ping(ctx); err != nil {
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

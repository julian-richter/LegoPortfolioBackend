package checks

import (
	"context"
	"time"

	"LegoManagerAPI/internal/api/handlers/health"
)

type ApplicationCheck struct{}

func NewApplicationCheck() *ApplicationCheck {
	return &ApplicationCheck{}
}

func (a *ApplicationCheck) Name() string {
	return "application"
}

func (a *ApplicationCheck) Check(ctx context.Context) health.Status {
	start := time.Now()

	return health.Status{
		Status:  "healthy",
		Latency: time.Since(start).String(),
	}
}

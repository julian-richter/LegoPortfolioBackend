package health

import (
	"context"
	"time"
)

// Status represents the health status of a service
type Status struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Response represents the overall health check result
type Response struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Environment string            `json:"environment"`
	Services    map[string]Status `json:"services"`
}

// Checker interface for individual health checks
type Checker interface {
	Name() string
	Check(ctx context.Context) Status
}

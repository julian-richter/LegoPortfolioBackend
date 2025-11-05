package health

import (
	"context"
	"sync"
	"time"
)

// Service orchestrates multiple health checks
type Service struct {
	checkers    []Checker
	environment string
}

// NewService creates a new health check service
func NewService(environment string, checkers ...Checker) *Service {
	return &Service{
		checkers:    checkers,
		environment: environment,
	}
}

// CheckAll runs all ehalth checks concurently
func (s *Service) CheckAll(ctx context.Context) Response {
	services := make(map[string]Status)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	// Run all checks concurrently
	for _, checker := range s.checkers {
		wg.Add(1)
		go func(checker Checker) {
			defer wg.Done()
			status := checker.Check(ctx)

			mu.Lock()
			services[checker.Name()] = status
			mu.Unlock()
		}(checker)
	}

	wg.Wait()

	// Determine the overall status
	overallStatus := "healthy"
	for _, status := range services {
		if status.Status != "healthy" {
			overallStatus = "unhealthy"
			break
		}
	}

	return Response{
		Status:      overallStatus,
		Timestamp:   time.Now().UTC(),
		Environment: s.environment,
		Services:    services,
	}
}

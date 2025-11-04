package handlers

import (
	"context"
	"net/http"
	"time"

	"LegoManagerAPI/internal/api/response"
	"LegoManagerAPI/internal/health"
)

type HealthHandler struct {
	healthService *health.Service
}

func NewHealthHandler(healthService *health.Service) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// Handle processes incoming health check requests, performs health checks, and sends a JSON response with the overall status.
func (h *HealthHandler) Handle(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	healthResponse := h.healthService.CheckAll(ctx)

	statusCode := http.StatusOK
	if healthResponse.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	response.JSON(res, statusCode, healthResponse)
}

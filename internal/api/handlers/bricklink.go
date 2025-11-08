package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"LegoManagerAPI/internal/api/response"
	"LegoManagerAPI/internal/api/service"
)

type BricklinkHandler struct {
	bricklinkService *service.BricklinkService
}

func NewBricklinkHandler(bricklinkService *service.BricklinkService) *BricklinkHandler {
	return &BricklinkHandler{
		bricklinkService: bricklinkService,
	}
}

// GetMinifig handles GET /api/bricklink/minifig/{id}
func (h *BricklinkHandler) GetMinifig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Extract minifig ID from path
	minifigID := strings.TrimPrefix(r.URL.Path, "/api/bricklink/minifig/")
	if minifigID == "" {
		response.Error(w, http.StatusBadRequest, "Minifig ID is required")
		return
	}

	// Fetch complete minifig data
	data, err := h.bricklinkService.GetMinifigComplete(ctx, minifigID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch minifig data: %v", err))
		return
	}

	// Convert to structured response
	structuredResponse := data.ToStructuredResponse()

	response.JSON(w, http.StatusOK, structuredResponse)
}

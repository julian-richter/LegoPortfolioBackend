package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"LegoManagerAPI/internal/api/dto"
	"LegoManagerAPI/internal/api/response"
	"LegoManagerAPI/internal/models"
	"LegoManagerAPI/internal/repos"
)

type UserHandler struct {
	userRepo *repos.UserRepository
}

func NewUserHandler(userRepo *repos.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate
	if req.Username == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// / Check if user already exists
	exists, err := h.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to check username existence")
		return
	}

	if exists {
		response.Error(w, http.StatusBadRequest, "Username already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to hash password")
	}

	// Create User
	user := &models.User{
		BaseModel:    models.BaseModel{},
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	response.JSON(w, http.StatusCreated, user)
}

// GetUser handles GET /api/users/:id
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	response.JSON(w, http.StatusOK, h.toUserResponse(user))
}

// UpdateUser handles PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Extract ID
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing user
	user, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	// Update fields
	user.Username = req.Username
	user.FirstName = req.FirstName
	user.LastName = req.LastName

	if err := h.userRepo.Update(ctx, user); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	response.JSON(w, http.StatusOK, h.toUserResponse(user))
}

// DeleteUser handles DELETE /api/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.userRepo.Delete(ctx, id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListUsers handles GET /api/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Parse query params
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := h.userRepo.List(ctx, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	total, err := h.userRepo.Count(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to count users")
		return
	}

	// Convert to response DTOs
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = h.toUserResponse(user)
	}

	resp := dto.ListUsersResponse{
		Users:  userResponses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	response.JSON(w, http.StatusOK, resp)
}

// SearchUsers handles GET /api/users/search?q=term
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		response.Error(w, http.StatusBadRequest, "Search term is required")
		return
	}

	users, err := h.userRepo.SearchByName(ctx, searchTerm)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = h.toUserResponse(user)
	}

	response.JSON(w, http.StatusOK, userResponses)
}

// UpdatePassword handles POST /api/users/{id}/password
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	idStr = strings.TrimSuffix(idStr, "/password")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user to verify old password
	user, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid old password")
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Update password
	if err := h.userRepo.UpdatePassword(ctx, id, string(newHash)); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper to convert model to response DTO
func (h *UserHandler) toUserResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		FullName:  user.FullName(), // Add this
		CreatedAt: user.CreatedAt,  // Add this
		UpdatedAt: user.UpdatedAt,  // Add this
	}
}

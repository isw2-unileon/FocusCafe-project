package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminCreateUserRequest defines the data needed by an admin to create a new user
type AdminCreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
}

// GetAllUsers returns a list of all registered users
func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.UserService.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUserByEmail searches for a user by their email address
// Query param: ?email=xxx
func (h *Handler) GetUserByEmail(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query parameter is required"})
		return
	}

	user, err := h.UserService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// AdminCreateUser allows an admin to create a new user with an assigned password.
// It follows the same flow as normal registration: Supabase Auth -> users table -> user_progress.
func (h *Handler) AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.Role = strings.TrimSpace(req.Role)

	if req.FirstName == "" || req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first name and last name are required"})
		return
	}
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters long"})
		return
	}
	if req.Role != "user" && req.Role != "admin" {
		req.Role = "user"
	}

	// 1. Create user in Supabase Auth
	userID, err := h.createAuthUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// 2. Create user profile in public.users
	if err := h.createUserProfile(userID, RegisterRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Create initial user progress
	if err := h.createUserProgress(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         userID,
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	})
}

// DeleteUser performs a hard delete of a user by ID.
// It first deletes from Supabase Auth (to free the email), then from the local database.
func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
		return
	}

	// 1. Delete from Supabase Auth (requires service_role key)
	if err := h.deleteAuthUser(id.String()); err != nil {
		slog.Error("failed to delete user from Supabase Auth, aborting", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete user from authentication provider: " + err.Error(),
		})
		return
	}

	// 2. Delete from local database (hard delete)
	if err := h.UserService.DeleteUser(c.Request.Context(), id); err != nil {
		slog.Error("Auth user deleted but local DB delete failed", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// deleteAuthUser removes a user from Supabase Auth using the service_role key.
// This frees the email address for reuse.
func (h *Handler) deleteAuthUser(userID string) error {
	if h.SupabaseServiceRoleKey == "" {
		return fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY not configured")
	}

	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/auth/v1/admin/users/%s", h.SupabaseURL, userID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.SupabaseServiceRoleKey))
	req.Header.Set("apikey", h.SupabaseServiceRoleKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Supabase Auth delete failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("supabase auth returned status %d: %s", resp.StatusCode, string(body))
	}

	slog.Info("user deleted from Supabase Auth", "id", userID)
	return nil
}
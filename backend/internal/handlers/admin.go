package handlers

import (
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

// DeleteUser performs a soft delete of a user by ID
func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
		return
	}

	if err := h.UserService.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}
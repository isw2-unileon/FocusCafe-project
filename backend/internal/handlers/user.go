package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
)

// getUserID is a helper function to extract and parse the user ID from the JWT claims in the context
func (h *Handler) getUserID(c *gin.Context) (uuid.UUID, error) {
	// Obtain user claims from context set by auth middleware
	claims, exists := c.Get("user")

	if !exists {
		return uuid.Nil, fmt.Errorf("user claims not found in context")
	}
	// Cast to UserClaims
	userClaims, ok := claims.(*auth.UserClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims format")
	}

	userID := userClaims.GetID()

	if userID == "" {
		return uuid.Nil, fmt.Errorf("empty user id")
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid format %w", err)
	}

	return id, nil
}

// GetUserProfile obtains the profile information of the authenticated user, including personal details and gamified stats (energy, level).
func (h *Handler) GetUserProfile(c *gin.Context) {
	// Obtain user ID from JWT claims
	id, _ := h.getUserID(c)

	// Obtain user profile from the service layer
	user, err := h.UserService.GetUserProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	// Return user profile as JSON response
	c.JSON(http.StatusOK, user)
}

// UpdateProfileRequest contains the data to update the profile.
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// GetLeaderboard returns the top 5 users ordered by experience (XP) descending.
func (h *Handler) GetLeaderboard(c *gin.Context) {
	// Verify the user is authenticated (middleware guarantees this, but we check for safety)
	if _, err := h.getUserID(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	profiles, err := h.UserService.GetLeaderboard(c.Request.Context(), 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}
	c.JSON(http.StatusOK, profiles)
}

// GetUserLeaderboardRank returns the current user's global rank and profile.
func (h *Handler) GetUserLeaderboardRank(c *gin.Context) {
	uid, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	rank, profile, err := h.UserService.GetUserLeaderboard(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user rank"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rank": rank,
		"user": profile,
	})
}

// UpdateUserProfile Updates the authenticated user's profile
// It only allows updating FirstName and LastName
func (h *Handler) UpdateUserProfile(c *gin.Context) {
	// Obtener userID del contexto usando el helper existente
	id, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse request body
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate that the fields are not empty (with trimming)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if req.FirstName == "" || req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first_name and last_name are required"})
		return
	}

	// Update user through the service layer
	if err := h.UserService.UpdateUserProfile(c.Request.Context(), id, req.FirstName, req.LastName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	// Get updated user with Progress through the service layer
	user, err := h.UserService.GetUserProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated profile"})
		return
	}

	// Returned the updated user
	c.JSON(http.StatusOK, user)
}

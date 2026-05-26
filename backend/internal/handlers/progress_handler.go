package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
)

// ProgressUpdateRequest defines the expected input from the frontend
type ProgressUpdateRequest struct {
	SessionID uint64 `json:"session_id" binding:"required"`
	Score     int    `json:"score" binding:"required,min=0"`
}

// UpdateProgressHandler updates user energy based on quiz results
func (h *Handler) UpdateProgressHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var input ProgressUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 1. Calculate rewards: 20 energy per correct answer
	energyEarned := input.Score * 20

	// 2. Use StudyService to perform database updates
	newTotal, err := h.StudyService.UpdateUserProgress(c.Request.Context(), userID, input.SessionID, energyEarned)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "study session not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Energy updated successfully",
		"energy_earned": energyEarned,
		"new_total":     newTotal,
	})
}

// getUserIDFromContext extracts and validates the UUID from Gin context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("user")
	if !exists {
		return uuid.Nil, errors.New("user identity not found")
	}
	claims, ok := val.(*auth.UserClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID format")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID format")
	}
	return userID, nil
}

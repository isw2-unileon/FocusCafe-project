package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

// ProgressUpdateRequest defines the expected input from the frontend
type ProgressUpdateRequest struct {
	SessionID uint64 `json:"session_id" binding:"required"`
	Score     int    `json:"score" binding:"min=0"`
}

// UpdateProgressHandler updates user energy based on quiz results
func UpdateProgressHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// 2. Perform database updates
		newTotal, err := applyProgressUpdate(db, userID, input.SessionID, energyEarned)
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
}

// getUserIDFromContext extracts and validates the UUID from Gin context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("user identity not found")
	}
	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID format")
	}
	return userID, nil
}

// applyProgressUpdate encapsulates the DB logic to reduce cognitive complexity
func applyProgressUpdate(db *gorm.DB, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
	var session models.StudySession
	if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return 0, errors.New("study session not found")
	}

	var progress models.UserProgress
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).First(&progress).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				progress = models.UserProgress{UserID: userID, Energy: 0, Level: 1, XP: 0}
				if err := tx.Create(&progress).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		progress.Energy += energy
		return tx.Save(&progress).Error
	})

	return progress.Energy, err
}

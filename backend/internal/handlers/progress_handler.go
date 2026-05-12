package handlers

import (
	"net/http"

	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProgressUpdateRequest defines the expected input from the frontend
type ProgressUpdateRequest struct {
	SessionID uint64 `json:"session_id" binding:"required"`
	Score     int    `json:"score" binding:"min=0"`
}

// UpdateProgressHandler updates user energy and experience based on quiz results
func UpdateProgressHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get UserID from context
		contextUserID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User identity not found"})
			return
		}

		userID, ok := contextUserID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		}

		var input ProgressUpdateRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// 1. Fetch the study session to get DurationMinutes
		var session models.StudySession
		if err := db.Where("id = ? AND user_id = ?", input.SessionID, userID).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Study session not found"})
			return
		}

		// 2. Calculate rewards
		// 20 per correct answer
		energyEarned := (input.Score * 10)

		// 3. Update or Create UserProgress
		var progress models.UserProgress
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ?", userID).First(&progress).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					progress = models.UserProgress{
						UserID: userID,
						Energy: 0,
						Level:  1,
						XP:     0,
					}
					if err := tx.Create(&progress).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			// Apply earned rewards
			progress.Energy += energyEarned

			// Level up logic: 100 XP per level
			if progress.XP >= 100 {
				levelsGained := progress.XP / 100
				progress.Level += levelsGained
				progress.XP = progress.XP % 100
			}

			return tx.Save(&progress).Error
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
			return
		}

		// 4. Return the results
		c.JSON(http.StatusOK, gin.H{
			"message":       "Energy updated successfully",
			"energy_earned": energyEarned,
			"new_total":     progress.Energy,
		})
	}
}

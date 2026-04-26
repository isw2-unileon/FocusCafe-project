package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

// StartStudySessionHandler initializes a new study session with a PDF
func StartStudySessionHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User credentials not found"})
		return
	}

	// Hacemos Type Assertion seguro para evitar que el servidor se caiga (panic)
	userMap, ok := userVal.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token format"})
		return
	}

	userIDStr, ok := userMap["sub"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Roken doesn't contain user ID (sub)"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID de usuario inválido en el token"})
		return
	}

	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "There is no PDF file!"})
		return
	}
	newFileName := uuid.New().String() + "_" + file.Filename
	filePath := "backend/uploads/" + newFileName

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while saving the file"})
		return
	}

	material := models.StudyMaterial{
		UserID:      userID,
		Title:       file.Filename,
		SubjectName: c.PostForm("subject_name"),
		FilePath:    filePath,
		UploadDate:  time.Now(),
	}

	if err := database.DB.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while registering study material."})
		return
	}

	session := models.StudySession{
		UserID:          userID,
		MaterialID:      material.ID,
		DurationMinutes: 25,
		StartTime:       time.Now(),
		Status:          "STUDYING",
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while creatin gstudy session."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id":  session.ID,
		"material_id": material.ID,
		"message":     "Let's study! Session has begun succesfully.",
	})
}

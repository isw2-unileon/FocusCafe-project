package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

// StartStudySessionHandler initializes a new study session, handles PDF upload, extracts text, and persists session data.
func StartStudySessionHandler(c *gin.Context) {
	log.Printf("--> Incoming study session start request")

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User credentials not found"})
		return
	}

	claims, ok := userVal.(*auth.UserClaims)
	if !ok {
		log.Printf("ERROR: User object in context is not of type *auth.UserClaims")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token format in context"})
		return
	}

	userIDStr := claims.Subject
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("ERROR: Invalid user ID format: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No PDF file provided"})
		return
	}

	// Generate a unique filename and set the storage path.
	newFileName := uuid.New().String() + "_" + file.Filename
	filePath := "backend/uploads/" + newFileName

	// Save the physical file to the local storage.
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Printf("ERROR SAVING FILE: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the uploaded file"})
		return
	}

	log.Printf("File successfully saved to: %s", filePath)

	// --- PDF TEXT EXTRACTION INTEGRATION ---
	// Extracting text content using the PDF service.
	content, err := services.ReadPdf(filePath)
	if err != nil {
		log.Printf("Error extracting text from PDF: %v", err)
		content = "" // Fallback to empty string to allow session creation to proceed.
	}

	// Create the study material record with the extracted content.
	material := models.StudyMaterial{
		UserID:      userID,
		Title:       file.Filename,
		SubjectName: c.PostForm("subject_name"),
		FilePath:    filePath,
		Content:     content,
		UploadDate:  time.Now(),
	}

	if err := database.DB.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while registering study material."})
		return
	}

	// Initialize the study session linked to the newly created material.
	session := models.StudySession{
		UserID:          userID,
		MaterialID:      material.ID,
		DurationMinutes: 25,
		StartTime:       time.Now(),
		Status:          "STUDYING",
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while creating study session."})
		return
	}

	// Logging session creation with proper type formatting.
	log.Printf("Session successfully created. ID: %d", session.ID)

	c.JSON(http.StatusCreated, gin.H{
		"session_id":  session.ID,
		"material_id": material.ID,
		"message":     "Let's study! Session has begun successfully.",
	})
}

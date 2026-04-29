package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth" // Asegúrate de que esta ruta es correcta
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

// StartStudySessionHandler initializes a new study session with a PDF
func StartStudySessionHandler(c *gin.Context) {
	log.Printf("--> Intento de inicio de sesión de estudio recibido")

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User credentials not found"})
		return
	}

	// CAMBIO CRÍTICO: El middleware guarda un puntero a auth.UserClaims
	claims, ok := userVal.(*auth.UserClaims)
	if !ok {
		log.Printf("ERROR: El tipo de usuario en el contexto no es *auth.UserClaims. Es: %T", userVal)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token format in context"})
		return
	}

	// Extraemos el ID usando la estructura de Claims
	userIDStr := claims.Subject
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("ERROR: ID de usuario no válido (%s): %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "There is no PDF file!"})
		return
	}

	// Generamos el nombre y la ruta
	newFileName := uuid.New().String() + "_" + file.Filename
	filePath := "backend/uploads/" + newFileName

	// Guardamos el archivo (Solo una vez)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Printf("ERROR AL GUARDAR ARCHIVO: %v | Ruta intentada: %s", err, filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while saving the file"})
		return
	}

	log.Printf("Archivo guardado con éxito en: %s", filePath)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while creating study session."})
		return
	}

	log.Printf("Sesión creada con éxito. ID: %s", session.ID)

	c.JSON(http.StatusCreated, gin.H{
		"session_id":  session.ID,
		"material_id": material.ID,
		"message":     "Let's study! Session has begun successfully.",
	})
}

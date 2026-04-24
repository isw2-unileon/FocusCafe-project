package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

func StartStudySessionHandler(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No se encontraron credenciales de usuario"})
		return
	}

	// Hacemos Type Assertion seguro para evitar que el servidor se caiga (panic)
	userMap, ok := userVal.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de token inválido"})
		return
	}

	userIDStr, ok := userMap["sub"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "El token no contiene el ID de usuario (sub)"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID de usuario inválido en el token"})
		return
	}

	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Archivo PDF requerido"})
		return
	}
	newFileName := uuid.New().String() + "_" + file.Filename
	filePath := "backend/uploads/" + newFileName

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el archivo en el servidor"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al registrar el material en la base de datos"})
		return
	}

	session := models.StudySession{
		UserID:          userID,
		MaterialID:      material.ID, // Usamos el ID que GORM acaba de generar para el material
		DurationMinutes: 25,
		StartTime:       time.Now(),
		Status:          "STUDYING",
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear la sesión de estudio"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id":  session.ID,
		"material_id": material.ID,
		"message":     "¡A estudiar! Sesión iniciada correctamente.",
	})
}

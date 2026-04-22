package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

// GetUserProfile obtiene el perfil completo del usuario autenticado
// Incluye información personal y estadísticas de progreso (energy, level)
func (h *Handler) GetUserProfile(c *gin.Context) {
	// Obtener claims del context (establecido por middleware Auth)
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user claims not found"})
		return
	}

	// Castear a UserClaims
	userClaims, ok := claims.(*auth.UserClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid claims format"})
		return
	}

	// Obtener userID del JWT (está en userClaims.ID)
	userID := userClaims.ID
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id missing from token"})
		return
	}

	// Parsear string a UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
		return
	}

	// Obtener usuario de la BD con relación Progress precargada
	var user models.User
	result := database.DB.Preload("Progress").First(&user, id)

	// Manejar errores de la BD
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Retornar usuario como JSON
	c.JSON(http.StatusOK, user)
}

// UpdateProfileRequest contiene los datos para actualizar el perfil
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UpdateUserProfile actualiza el perfil del usuario autenticado
// Solo permite actualizar FirstName y LastName
func (h *Handler) UpdateUserProfile(c *gin.Context) {
	// Obtener claims del context
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user claims not found"})
		return
	}

	// Castear a UserClaims
	userClaims, ok := claims.(*auth.UserClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid claims format"})
		return
	}

	// Obtener userID
	userID := userClaims.ID
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id missing from token"})
		return
	}

	// Parsear string a UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
		return
	}

	// Parsear request body
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validar que los campos no estén vacíos (con trimming)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if req.FirstName == "" || req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first_name and last_name are required"})
		return
	}

	// Actualizar usuario en la BD
	var user models.User
	result := database.DB.Model(&user).Where("id = ?", id).Updates(models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	// Obtener usuario actualizado con Progress
	database.DB.Preload("Progress").First(&user, id)

	// Retornar usuario actualizado
	c.JSON(http.StatusOK, user)
}

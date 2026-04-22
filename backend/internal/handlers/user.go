package handlers

import (
	"net/http"

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

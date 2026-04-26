package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
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
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	// Return user profile as JSON response
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

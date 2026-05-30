package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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
	// Obtener userID del contexto usando el helper existente
	id, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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

	// Actualizar usuario a través del service layer
	if err := h.UserService.UpdateUserProfile(c.Request.Context(), id, req.FirstName, req.LastName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	// Obtener usuario actualizado con Progress a través del service layer
	user, err := h.UserService.GetUserProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated profile"})
		return
	}

	// Retornar usuario actualizado
	c.JSON(http.StatusOK, user)
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateQuizFromSession Handles the AI quiz generation request for a given session
func CreateQuizFromSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	c.JSON(http.StatusOK, gin.H{
		"message":    "IA lista para generar el quiz",
		"session_id": sessionID,
	})
}

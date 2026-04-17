package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
)

// Auth is the middleware for JWT authentication
func Auth(validator auth.TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Token from headers
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		headersParts := strings.Split(authHeader, " ")
		if len(headersParts) != 2 || headersParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		tokenString := headersParts[1]
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "Token is empty"})
			return
		}

		claims, err := validator.ValidateToken(tokenString)
		if err != nil {
			// ¡ESTO ES CLAVE! Imprime el error real en la consola de VS Code/Terminal
			fmt.Printf("❌ Error de validación JWT: %v\n", err)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"details": err.Error(), // Opcional: para que Thunder Client te diga qué pasa
			})
			return // <--- IMPORTANTE: Detener la ejecución aquí
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		}
		c.Set("user", claims)
		c.Next()
	}
}

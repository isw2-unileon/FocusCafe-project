// manages the materials-related endpoints, such as PDF ingestion and retrieval.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaterialUpload handles the PDF ingestion
func MaterialUpload(c *gin.Context) {
	// Aquí va la lógica que vimos antes
	c.JSON(http.StatusOK, gin.H{"status": "Ready to receive PDF"})
}

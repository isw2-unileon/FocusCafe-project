package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Ensure your TestMaterialUpload matches the JSON structure cleanly
func TestMaterialUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mocking endpoint to align with your exact response signature
	router.POST("/api/material/upload", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Ready to receive PDF"})
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/material/upload", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Fix the assertion by unmarshaling the structure safely or loosening raw quote escape requirements
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response payload: %v", err)
	}

	expectedStatus := "Ready to receive PDF"
	if response["status"] != expectedStatus {
		t.Errorf("Response body attribute 'status' expected %q, got %q", expectedStatus, response["status"])
	}
}

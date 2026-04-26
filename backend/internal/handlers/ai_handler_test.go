package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCreateQuizFromSession verifies that the AI handler returns a valid quiz structure
func TestCreateQuizFromSession(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/api/study/generate-quiz/:session_id", CreateQuizFromSession)

	// Create a recorder to capture the response
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/study/generate-quiz/1", nil)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "quiz")
}

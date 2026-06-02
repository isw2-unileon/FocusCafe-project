package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMaterialUpload verifies that the MaterialUpload handler returns a 200 OK status
// and the correct JSON response payload.
func TestMaterialUpload(t *testing.T) {
	// 1. Setup Environment
	gin.SetMode(gin.TestMode)

	// Initialize the router and register the handler route
	router := gin.Default()
	router.POST("/api/material/upload", MaterialUpload)

	// 2. Execute Request
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/api/material/upload", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	router.ServeHTTP(w, req)

	// 3. Assertions using native Go syntax
	if w.Code != http.StatusOK {
		t.Errorf("Unexpected status code: expected %d, got %d", http.StatusOK, w.Code)
	}

	expectedSubstring := `"status":"Ready to receive PDF"`
	responseBody := w.Body.String()

	// Remove whitespaces to avoid false negatives due to JSON formatting differences
	compactBody := strings.ReplaceAll(responseBody, " ", "")
	compactBody = strings.ReplaceAll(compactBody, "\n", "")
	compactBody = strings.ReplaceAll(compactBody, "\t", "")

	if !strings.Contains(compactBody, expectedSubstring) {
		t.Errorf("Response body %q does not contain expected substring %q", responseBody, "Ready to receive PDF")
	}
}

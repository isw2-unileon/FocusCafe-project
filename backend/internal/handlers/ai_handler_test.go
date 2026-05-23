package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite database for testing purposes.
func setupTestDB() {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&models.StudySession{}, &models.StudyMaterial{})
	database.DB = db
}

type mockAIService struct{}

func (m *mockAIService) GenerateQuiz(pdfText string) (string, error) {
	return `{"quiz_name": "Test Quiz", "questions": []}`, nil
}

// TestCreateQuizFromSession verifies that the handler correctly processes a session and attempts AI generation.
func TestCreateQuizFromSession(t *testing.T) {
	// 1. Setup Environment
	gin.SetMode(gin.TestMode)
	setupTestDB()

	mockAI := &mockAIService{}
	h := &Handler{
		AIService: mockAI,
	}

	router := gin.Default()
	router.POST("/api/study/generate-quiz/:session_id", h.CreateQuizFromSession)

	// 2. Seed Mock Data
	userID := uuid.New()
	material := models.StudyMaterial{
		UserID:  userID,
		Content: "This is a test content about Software Engineering.",
	}
	database.DB.Create(&material)

	session := models.StudySession{
		ID:         1,
		UserID:     userID,
		MaterialID: material.ID,
	}
	database.DB.Create(&session)

	// 3. Execute Request
	w := httptest.NewRecorder()
	// Using the ID we just created
	req, err := http.NewRequest("POST", "/api/study/generate-quiz/1", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	router.ServeHTTP(w, req)

	// 4. Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Quiz")
}

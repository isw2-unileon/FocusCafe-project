package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite database for testing purposes.
func setupTestDB() {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Error opening test database: " + err.Error())
	}

	err = database.DB.AutoMigrate(&models.StudySession{}, &models.StudyMaterial{}, &models.Quiz{})
	if err != nil {
		panic("Error running test migrations: " + err.Error())
	}
}

type mockAIService struct{}

// GenerateQuiz returns a fixed quiz JSON string for testing purposes.
func (m *mockAIService) GenerateQuiz(pdfText string) (string, error) {
	return `{"quiz_name": "Test Quiz", "questions": []}`, nil
}

// TestCreateQuizFromSession verifies that the handler correctly processes a valid session and returns a 200 OK status.
func TestCreateQuizFromSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB()

	studyRepo := repository.NewStudyRepository(database.DB)
	studyService := services.NewStudyService(studyRepo)
	mockAI := &mockAIService{}

	h := &Handler{
		AIService:    mockAI,
		StudyService: studyService,
	}

	router := gin.Default()
	router.POST("/api/study/generate-quiz/:session_id", h.CreateQuizFromSession)

	userID := uuid.New()
	material := models.StudyMaterial{
		UserID:  userID,
		Title:   "Software Engineering Notes",
		Content: "This is a test content about Software Engineering.",
	}
	database.DB.Create(&material)

	session := models.StudySession{
		ID:         1,
		UserID:     userID,
		MaterialID: material.ID,
		Status:     "STUDYING",
		StartTime:  time.Now(),
	}
	database.DB.Create(&session)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/api/study/generate-quiz/1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	router.ServeHTTP(w, req)

	// 4. Assertions using native Go
	if w.Code != http.StatusOK {
		t.Errorf("Unexpected status code: expected %d, got %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Test Quiz") {
		t.Errorf("Response body %q does not contain expected substring %q", w.Body.String(), "Test Quiz")
	}
}

// TestCreateQuizFromSession_NotFound verifies that the handler handles an invalid or non-existent session ID properly.
func TestCreateQuizFromSession_NotFound(t *testing.T) {
	// 1. Setup Environment
	gin.SetMode(gin.TestMode)
	setupTestDB()

	studyRepo := repository.NewStudyRepository(database.DB)
	studyService := services.NewStudyService(studyRepo)
	mockAI := &mockAIService{}

	h := &Handler{
		AIService:    mockAI,
		StudyService: studyService,
	}

	router := gin.Default()
	router.POST("/api/study/generate-quiz/:session_id", h.CreateQuizFromSession)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/api/study/generate-quiz/9999", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("Expected error status code, but got %d", w.Code)
	}
}

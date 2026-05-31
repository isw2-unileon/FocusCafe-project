package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite database for testing purposes.
func setupTestDB() {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&models.StudySession{}, &models.StudyMaterial{})
	database.DB = db
}

// mockAIService is a simple mock implementation of the AIServiceInterface for testing.
type mockAIService struct{}

// GenerateQuiz returns a fixed quiz JSON string for testing purposes.
func (m *mockAIService) GenerateQuiz(pdfText string) (string, error) {
	return `{"quiz_name": "Test Quiz", "questions": []}`, nil
}

// TestCreateQuizFromSession verifies that the handler correctly processes a session and attempts AI generation.
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
	req, _ := http.NewRequest("POST", "/api/study/generate-quiz/1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Quiz")
}

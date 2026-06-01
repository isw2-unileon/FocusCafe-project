package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

// MockAIService implements AIServiceInterface for testing
type MockAIService struct{}

func (m *MockAIService) GenerateQuiz(content string) (string, error) {
	return `{"quiz": "test quiz from mock"}`, nil
}

func TestStudy_Integration(t *testing.T) {
	db, h := setupTestApp()
	h.AIService = &MockAIService{}

	t.Run("Start Study Session - Success", func(t *testing.T) {
		userID := uuid.New()
		seedUserProgressData(db, userID)
		testStartStudySessionSuccess(t, db, h, userID)
	})

	t.Run("Create Quiz from Session - Success", func(t *testing.T) {
		userID := uuid.New()
		sessionID := seedStudyTestData(db, userID)
		testCreateQuizSuccess(t, h, userID, sessionID)
	})

	t.Run("Update Progress - Success", func(t *testing.T) {
		userID := uuid.New()
		sessionID := seedStudyTestData(db, userID)

		router := setupStudyRouter(h, userID)
		testUpdateProgressSuccess(t, router, db, userID, sessionID)
	})

	t.Run("Update Progress - Invalid Session ID", func(t *testing.T) {
		userID := uuid.New()
		_ = seedStudyTestData(db, userID)

		router := setupStudyRouter(h, userID)
		testUpdateProgressInvalidSession(t, router)
	})

	t.Run("Update Progress - Missing Body Fields", func(t *testing.T) {
		userID := uuid.New()
		_ = seedStudyTestData(db, userID)

		router := setupStudyRouter(h, userID)
		testUpdateProgressMissingFields(t, router)
	})
}

func seedUserProgressData(db *gorm.DB, userID uuid.UUID) {
	db.Create(&models.UserProgress{
		UserID: userID,
		Level:  1,
		Energy: 100,
		XP:     0,
	})
}

func testStartStudySessionSuccess(t *testing.T, db *gorm.DB, h *handlers.Handler, userID uuid.UUID) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("pdf", "test.pdf")
	_, _ = part.Write([]byte("%PDF-1.4 test content"))
	_ = writer.WriteField("subject_name", "Math")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/study/start", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	c, _ := setupSubtestContext(w, req, userID)
	h.StartStudySessionHandler(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal study quiz content: %v", err)
	}
	sessionID := uint64(resp["session_id"].(float64))

	// Verify DB
	var session models.StudySession
	if err := db.Preload("Material").First(&session, sessionID).Error; err != nil {
		t.Fatalf("failed to find session in DB: %v", err)
	}
	if session.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, session.UserID)
	}
}

func testCreateQuizSuccess(t *testing.T, h *handlers.Handler, userID uuid.UUID, sessionID uint64) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/study/generate-quiz/%d", sessionID), nil)

	c, _ := setupSubtestContext(w, req, userID)
	c.Params = []gin.Param{{Key: "session_id", Value: fmt.Sprintf("%d", sessionID)}}
	h.CreateQuizFromSession(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("failed to unmarshal generated quiz response payload: %v", err)
	}

	quizContent, exists := responseBody["quiz"]
	if !exists {
		if !strings.Contains(w.Body.String(), "test quiz from mock") {
			t.Errorf("expected quiz content in response, got %s", w.Body.String())
		}
	} else if !strings.Contains(fmt.Sprintf("%v", quizContent), "test quiz from mock") {
		t.Errorf("expected quiz content inside verified JSON, got %v", quizContent)
	}
}

func setupStudyRouter(h *handlers.Handler, userID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	api.Use(mockAuthMiddleware(userID)) // Injects the target user authenticated context safely
	api.POST("/user/progress", h.UpdateProgressHandler)

	return r
}

func seedStudyTestData(db *gorm.DB, userID uuid.UUID) uint64 {
	db.Create(&models.UserProgress{
		UserID: userID,
		Level:  1,
		Energy: 100,
		XP:     0,
	})

	material := models.StudyMaterial{
		UserID:      userID,
		Title:       "Test Material",
		SubjectName: "Math",
		Content:     "Some PDF text content",
		UploadDate:  time.Now(),
	}
	db.Create(&material)

	session := models.StudySession{
		UserID:     userID,
		MaterialID: material.ID,
		Status:     "active",
		StartTime:  time.Now(),
	}
	db.Create(&session)

	return session.ID
}

func testUpdateProgressSuccess(t *testing.T, r *gin.Engine, db *gorm.DB, userID uuid.UUID, sessionID uint64) {
	// Score = 3 -> 3 * 20 = 60 energy
	reqBody, _ := json.Marshal(map[string]interface{}{
		"session_id": sessionID,
		"score":      3,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(reqBody))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal updated progress result: %v", err)
	}
	if response["energy_earned"].(float64) != 60 {
		t.Errorf("expected energy_earned 60, got %v", response["energy_earned"])
	}
	if response["new_total"].(float64) != 160 {
		t.Errorf("expected new_total 160, got %v", response["new_total"])
	}

	// Verify DB
	var progress models.UserProgress
	db.Where("user_id = ?", userID).First(&progress)
	if progress.Energy != 160 {
		t.Errorf("DB: expected energy 160, got %d", progress.Energy)
	}
}

func testUpdateProgressInvalidSession(t *testing.T, r *gin.Engine) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"session_id": 9999,
		"score":      1,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(reqBody))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func testUpdateProgressMissingFields(t *testing.T, r *gin.Engine) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"score": 1,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(reqBody))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

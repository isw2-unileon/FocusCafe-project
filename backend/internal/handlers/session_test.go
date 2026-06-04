package handlers

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
	"gorm.io/gorm"
)

// Global mock user ID used across tests
var mockUserID = uuid.NewString()

// setupSessionTest initializes the database and dependencies before each test case
func setupSessionTest(t *testing.T) *Handler { //
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error opening in-memory database: %v", err)
	}

	err = db.AutoMigrate(&models.StudyMaterial{}, &models.StudySession{})
	if err != nil {
		t.Fatalf("Error running test migrations: %v", err)
	}

	database.DB = db

	studyRepo := repository.NewStudyRepository(db)
	studyService := services.NewStudyService(studyRepo)
	handler := &Handler{
		StudyService: studyService,
	}

	_ = os.MkdirAll("backend/uploads", 0o750)

	return handler
}

// teardownSessionTest cleans up temporary folders and assets created during tests
func teardownSessionTest() {
	_ = os.RemoveAll("backend")
}

// TestStartStudySessionSuccess verifies the successful creation of a study session.
func TestStartStudySessionSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupSessionTest(t)
	defer teardownSessionTest()

	recorder := httptest.NewRecorder()

	mockClaims := &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: mockUserID,
		},
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("pdf", "test_material.pdf")
	if _, err := part.Write([]byte("fake pdf content")); err != nil {
		t.Fatalf("Failed to write part: %v", err)
	}
	if err := writer.WriteField("subject_name", "Software Engineering"); err != nil {
		t.Fatalf("Failed to write field: %v", err)
	}
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/study/start", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, r := gin.CreateTestContext(recorder)
	r.POST("/api/study/start", func(c *gin.Context) {
		c.Set("user", mockClaims)
		handler.StartStudySessionHandler(c)
	})

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Unexpected status code: expected %d, got %d", http.StatusCreated, recorder.Code)
	}

	responseBody := recorder.Body.String()
	if !strings.Contains(responseBody, "session_id") {
		t.Errorf("Response body does not contain 'session_id'. Body: %s", responseBody)
	}
	if !strings.Contains(responseBody, "material_id") {
		t.Errorf("Response body does not contain 'material_id'. Body: %s", responseBody)
	}
}

// TestStartStudySessionNoFile verifies that the handler fails with a 400 status when no PDF is provided.
func TestStartStudySessionNoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupSessionTest(t)
	defer teardownSessionTest()

	recorder := httptest.NewRecorder()
	_, r := gin.CreateTestContext(recorder)

	mockClaims := &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: mockUserID,
		},
	}

	r.POST("/api/study/start", func(c *gin.Context) {
		c.Set("user", mockClaims)
		handler.StartStudySessionHandler(c)
	})

	req, _ := http.NewRequest("POST", "/api/study/start", nil)
	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Unexpected status code: expected %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

// TestStartStudySessionNoUserInContext verifies the handler fails with 401 when user is missing in context.
func TestStartStudySessionNoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupSessionTest(t)
	defer teardownSessionTest()

	recorder := httptest.NewRecorder()
	_, r := gin.CreateTestContext(recorder)

	r.POST("/api/study/start", handler.StartStudySessionHandler)

	req, _ := http.NewRequest("POST", "/api/study/start", nil)
	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", recorder.Code)
	}
}

// TestStartStudySessionInvalidUserType verifies the handler fails with 500 when user in context is of wrong type.
func TestStartStudySessionInvalidUserType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupSessionTest(t)
	defer teardownSessionTest()

	recorder := httptest.NewRecorder()
	_, r := gin.CreateTestContext(recorder)

	r.POST("/api/study/start", func(c *gin.Context) {
		c.Set("user", "not a UserClaims object")
		handler.StartStudySessionHandler(c)
	})

	req, _ := http.NewRequest("POST", "/api/study/start", nil)
	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}
}

// TestStartStudySessionInvalidUserID verifies the handler fails with 500 when user ID in token is invalid.
func TestStartStudySessionInvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupSessionTest(t)
	defer teardownSessionTest()

	recorder := httptest.NewRecorder()
	_, r := gin.CreateTestContext(recorder)

	mockClaims := &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "invalid-uuid",
		},
	}

	r.POST("/api/study/start", func(c *gin.Context) {
		c.Set("user", mockClaims)
		handler.StartStudySessionHandler(c)
	})

	req, _ := http.NewRequest("POST", "/api/study/start", nil)
	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}
}

// TestStartStudySessionSaveFileError verifies the handler fails with 500 when file saving fails.
func TestStartStudySessionSaveFileError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupSessionTest(t)
	defer teardownSessionTest()

	// Make the uploads directory un-writable or non-existent to force an error
	os.RemoveAll("backend")
	// Create a file named 'backend' so MkdirAll and file creation inside it fails
	if err := os.WriteFile("backend", []byte("blocker"), 0644); err != nil {
		t.Fatalf("Failed to create blocker file: %v", err)
	}

	recorder := httptest.NewRecorder()

	mockClaims := &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: mockUserID,
		},
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("pdf", "test.pdf")
	if _, err := part.Write([]byte("content")); err != nil {
		t.Fatalf("Failed to write part: %v", err)
	}
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/study/start", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, r := gin.CreateTestContext(recorder)
	r.POST("/api/study/start", func(c *gin.Context) {
		c.Set("user", mockClaims)
		handler.StartStudySessionHandler(c)
	})

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}
}

// MockStudyService implements services.StudyServiceInterface for testing purposes.
type MockStudyService struct {
	services.StudyServiceInterface
	StartStudySessionFunc func(ctx context.Context, userID uuid.UUID, fileName, subjectName, filePath, content string) (*domain.StudySession, uint64, error)
}

func (m *MockStudyService) StartStudySession(ctx context.Context, userID uuid.UUID, fileName, subjectName, filePath, content string) (*domain.StudySession, uint64, error) {
	return m.StartStudySessionFunc(ctx, userID, fileName, subjectName, filePath, content)
}

// TestStartStudySessionServiceError verifies that the handler handles service errors correctly.
func TestStartStudySessionServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupSessionTest(t)
	defer teardownSessionTest()

	// Inject the mock service
	mockService := &MockStudyService{
		StartStudySessionFunc: func(_ context.Context, _ uuid.UUID, _, _, _, _ string) (*domain.StudySession, uint64, error) {
			return nil, 0, errors.New("simulated service error")
		},
	}
	handler.StudyService = mockService

	recorder := httptest.NewRecorder()

	mockClaims := &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: mockUserID,
		},
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("pdf", "test.pdf")
	if _, err := part.Write([]byte("content")); err != nil {
		t.Fatalf("Failed to write part: %v", err)
	}
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/study/start", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, r := gin.CreateTestContext(recorder)
	r.POST("/api/study/start", func(c *gin.Context) {
		c.Set("user", mockClaims)
		handler.StartStudySessionHandler(c)
	})

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Unexpected status code: expected %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "Error while registering study material or session.") {
		t.Errorf("Expected error message not found in response")
	}
}

package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SessionTestSuite defines the suite for session handler testing
type SessionTestSuite struct {
	suite.Suite
	db *gorm.DB
}

// SetupSuite initializes the in-memory database and required folders
func (suite *SessionTestSuite) SetupSuite() {
	var err error
	suite.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)

	// Migrate models to in-memory database
	err = suite.db.AutoMigrate(&models.StudyMaterial{}, &models.StudySession{})
	assert.NoError(suite.T(), err)

	// Inject the mock DB into the global database instance
	database.DB = suite.db

	// Create temporary uploads folder for tests
	_ = os.MkdirAll("backend/uploads", 0o750)
}

// TearDownSuite cleans up the temporary files after tests
func (suite *SessionTestSuite) TearDownSuite() {
	_ = os.RemoveAll("backend") // Cleanup the fake uploads folder
}

// TestStartStudySessionSuccess verifies the successful creation of a study session
func (suite *SessionTestSuite) TestStartStudySessionSuccess() {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, r := gin.CreateTestContext(recorder)

	// Mock user data in context
	userID := uuid.New().String()
	ctx.Set("user", map[string]interface{}{
		"sub": userID,
	})

	// Prepare multipart form with a fake PDF file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("pdf", "test_material.pdf")
	part.Write([]byte("fake pdf content"))
	writer.WriteField("subject_name", "Software Engineering")
	writer.Close()

	// Setup request and router
	req, _ := http.NewRequest("POST", "/api/study/start", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	r.POST("/api/study/start", func(c *gin.Context) {
		// Pass the mocked user to the context since r.ServeHTTP creates a new one
		c.Set("user", ctx.MustGet("user"))
		StartStudySessionHandler(c)
	})

	r.ServeHTTP(recorder, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusCreated, recorder.Code)
	assert.Contains(suite.T(), recorder.Body.String(), "session_id")
	assert.Contains(suite.T(), recorder.Body.String(), "material_id")
}

// TestStartStudySessionNoFile verifies that the handler fails when no PDF is provided
func (suite *SessionTestSuite) TestStartStudySessionNoFile() {
	recorder := httptest.NewRecorder()
	_, r := gin.CreateTestContext(recorder)

	userID := uuid.New().String()
	r.POST("/api/study/start", func(c *gin.Context) {
		c.Set("user", map[string]interface{}{"sub": userID})
		StartStudySessionHandler(c)
	})

	// Request without files
	req, _ := http.NewRequest("POST", "/api/study/start", nil)
	r.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
}

// Run the test suite
func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}

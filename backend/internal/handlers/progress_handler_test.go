package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUpdateProgressHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Setup SQL Mock
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock: %s", err)
	}
	defer sqlDB.Close()

	gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})

	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func()
		wantStatusCode int
		expectedBody   string
	}{
		{
			name: "Success: Earn 30 energy (3 correct answers)",
			requestBody: gin.H{
				"session_id": 1,
				"score":      3,
			},
			setupMock: func() {
				// Mocking the StudySession check
				mock.ExpectQuery(`SELECT \* FROM "study_sessions"`).
					WithArgs(1, userID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, userID))

				// Mocking the UserProgress fetch
				mock.ExpectQuery(`SELECT \* FROM "user_progress"`).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"user_id", "energy"}).AddRow(userID, 10))

				// Mocking the Save (Update) - We expect 10 (old) + 30 (new) = 40
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "user_progress"`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"energy_earned":30`,
		},
		{
			name: "Error: Session Not Found",
			requestBody: gin.H{
				"session_id": 99,
				"score":      3,
			},
			setupMock: func() {
				mock.ExpectQuery(`SELECT \* FROM "study_sessions"`).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			expectedBody:   `"error":"Study session not found"`,
		},
		{
			name: "Error: Invalid Body",
			requestBody: gin.H{
				"score": -1, // Should trigger validation error
			},
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `"error":"Invalid request body"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// 2. Setup Gin Recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Simulate Authentication Context
			c.Set("userID", userID)

			// Setup Request Body
			jsonBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// 3. Execute Handler
			handler := handlers.UpdateProgressHandler(gormDB)
			handler(c)

			// 4. Assertions
			if w.Code != tt.wantStatusCode {
				t.Errorf("%s: status = %v, want %v", tt.name, w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("%s: body = %v, want to contain %v", tt.name, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

// mockProgressStudyService is a test double for StudyServiceInterface used in progress tests.
type mockProgressStudyService struct {
	updateUserProgressFunc func(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error)
}

func (m *mockProgressStudyService) StartStudySession(ctx context.Context, userID uuid.UUID, fileName, subjectName, filePath, content string) (*domain.StudySession, uint64, error) {
	return nil, 0, errors.New("not mocked")
}

func (m *mockProgressStudyService) GetSessionWithMaterial(ctx context.Context, sessionID uint64) (*models.StudySession, error) {
	return nil, errors.New("not mocked")
}

func (m *mockProgressStudyService) SaveQuiz(ctx context.Context, sessionID uint64, quizJSON string) error {
	return errors.New("not mocked")
}

func (m *mockProgressStudyService) UpdateUserProgress(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
	if m.updateUserProgressFunc != nil {
		return m.updateUserProgressFunc(ctx, userID, sessionID, energy)
	}
	return 0, errors.New("updateUserProgress not mocked")
}

func TestUpdateProgressHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name             string
		body             string
		userIDInContext  uuid.UUID
		mockBehavior     func(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error)
		wantStatusCode   int
		expectedBodyPart string
	}{
		{
			name:            "Success: Earn 60 energy with 3 correct answers",
			body:            `{"session_id":1,"score":3}`,
			userIDInContext: userID,
			mockBehavior: func(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
				return 70, nil // 10 existing + 60 earned
			},
			wantStatusCode:   http.StatusOK,
			expectedBodyPart: `"energy_earned":60`,
		},
		{
			name:            "Error: Session not found returns 404",
			body:            `{"session_id":99,"score":2}`,
			userIDInContext: userID,
			mockBehavior: func(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
				return 0, errors.New("study session not found")
			},
			wantStatusCode:   http.StatusNotFound,
			expectedBodyPart: `"error":"study session not found"`,
		},
		{
			name:            "Error: Missing user in context returns 401",
			body:            `{"session_id":1,"score":3}`,
			userIDInContext: uuid.Nil,
			mockBehavior:    nil,
			wantStatusCode:  http.StatusUnauthorized,
			expectedBodyPart: `"error"`,
		},
		{
			name:            "Error: Invalid request body returns 400",
			body:            `{"session_id":1}`,
			userIDInContext: userID,
			mockBehavior:    nil,
			wantStatusCode:  http.StatusBadRequest,
			expectedBodyPart: `"error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockProgressStudyService{updateUserProgressFunc: tt.mockBehavior}
			h := &handlers.Handler{StudyService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				claims := &auth.UserClaims{
					RegisteredClaims: jwt.RegisteredClaims{Subject: tt.userIDInContext.String()},
				}
				c.Set("user", claims)
			}

			c.Request = httptest.NewRequest("POST", "/api/user/progress", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.UpdateProgressHandler(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.UpdateProgressHandler() status = %v, want %v. Body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}

			if !strings.Contains(w.Body.String(), tt.expectedBodyPart) {
				t.Errorf("Handler.UpdateProgressHandler() body = %v, want to contain %v", w.Body.String(), tt.expectedBodyPart)
			}
		})
	}
}

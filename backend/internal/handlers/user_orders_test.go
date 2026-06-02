package handlers_test

import (
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
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/ws"
)

type mockUserOrdersService struct {
	getUserOrdersFunc func(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error)
	completeUserOrder func(ctx context.Context, userId uuid.UUID, orderId uint) error
}

func (m *mockUserOrdersService) GetUserOrders(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
	return m.getUserOrdersFunc(ctx, id)
}

func (m *mockUserOrdersService) CompleteUserOrder(ctx context.Context, userID uuid.UUID, orderID uint) error {
	return m.completeUserOrder(ctx, userID, orderID)
}

func TestHandler_GetUserOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		userIDInContext uuid.UUID
		mockBehavior    func(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error)
		wantStatusCode  int
		expectedBody    string
	}{
		{
			name:            "Success: Returns 200 and Orders",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
				return []domain.UserOrder{
					{ID: 1, Status: "pending"},
				}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"status":"pending"`,
		},
		{
			name:            "Error: Service failure returns 500",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
				return nil, errors.New("database error")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to get user orders"}`,
		},
		{
			name:            "Error: No user in context returns 401",
			userIDInContext: uuid.Nil,
			mockBehavior:    nil, // Should not be called
			wantStatusCode:  http.StatusUnauthorized,
			expectedBody:    `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockUserOrdersService{getUserOrdersFunc: tt.mockBehavior}
			h := &handlers.Handler{UserOrdersService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				claims := &auth.UserClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: tt.userIDInContext.String(),
					},
				}
				c.Set("user", claims)
			}

			c.Request = httptest.NewRequest("GET", "/api/v1/users/me/orders", nil)

			h.GetUserOrders(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.GetUserOrders() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.GetUserOrders() body = %v, want to contain %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestHandler_CompleteUserOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		orderIDParam    string
		userIDInContext uuid.UUID
		mockBehavior    func(ctx context.Context, userId uuid.UUID, orderId uint) error
		wantStatusCode  int
		expectedBody    string
		enableWS        bool
		forceUserSvcErr bool
	}{
		{
			name:            "Success: Completes order and returns 200",
			orderIDParam:    "123",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, userId uuid.UUID, orderId uint) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"message":"Order successfully completed!"`,
			enableWS:       true,
		},
		{
			name:            "Error: Invalid order ID returns 400",
			orderIDParam:    "abc",
			userIDInContext: uuid.New(),
			mockBehavior:    nil,
			wantStatusCode:  http.StatusBadRequest,
			expectedBody:    `{"error":"Invalid order ID"}`,
			enableWS:        false,
		},
		{
			name:            "Error: Insufficient energy returns 400",
			orderIDParam:    "123",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, userId uuid.UUID, orderId uint) error {
				return errors.New("insufficient energy")
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"Not enough energy"}`,
			enableWS:       false,
		},
		{
			name:            "Error: Service failure returns 500",
			orderIDParam:    "123",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, userId uuid.UUID, orderId uint) error {
				return errors.New("database error")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `{"error":"Error at completing the order: database error"}`,
			enableWS:       false,
		},
		{
			name:            "Error: No user in context returns 401",
			orderIDParam:    "123",
			userIDInContext: uuid.Nil,
			mockBehavior:    nil,
			wantStatusCode:  http.StatusUnauthorized,
			expectedBody:    `{"error":"unauthorized"}`,
			enableWS:        true,
		},
		{
			name:            "Error: Order already completed returns 409 Conflict",
			orderIDParam:    "123",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, userId uuid.UUID, orderId uint) error {
				return errors.New("order already completed")
			},
			wantStatusCode: http.StatusConflict,
			expectedBody:   `"error":"This order has already been completed by another user"`,
			enableWS:       true,
		},
		{
			name:            "Error: WebSocket broadcast fails due to user service error",
			orderIDParam:    "123",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, userId uuid.UUID, orderId uint) error {
				return nil // Order completes successfully
			},
			wantStatusCode:  http.StatusOK,
			enableWS:        true,
			forceUserSvcErr: true,
		},
		{
			name:            "Success: Completes order but WS is disabled",
			orderIDParam:    "125",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, userId uuid.UUID, orderId uint) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"message":"Order successfully completed!"`,
			enableWS:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockUserOrdersService{completeUserOrder: tt.mockBehavior}

			// Initialize Mocks
			spyWS := &ws.Hub{}
			mUser := &mockUserService{
				getUserProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
					if tt.forceUserSvcErr {
						return nil, errors.New("simulated database error")
					}
					return &domain.UserProfile{
						Role: "user",
						Group: &domain.Group{
							ID: 12345,
						},
					}, nil
				},
			}
			h := &handlers.Handler{
				UserOrdersService: mService,
				UserService:       mUser,
			}

			if tt.enableWS {
				h.WSHub = spyWS
			} else {
				h.WSHub = nil // Explicitly ensure it's nil for the "else" branch
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				claims := &auth.UserClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: tt.userIDInContext.String(),
					},
				}
				c.Set("user", claims)
			}

			c.Params = gin.Params{{Key: "id", Value: tt.orderIDParam}}
			c.Request = httptest.NewRequest("POST", "/api/v1/users/me/orders/"+tt.orderIDParam+"/complete", nil)

			h.CompleteUserOrder(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.CompleteUserOrder() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.expectedBody != "" && !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.CompleteUserOrder() body = %v, want to contain %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

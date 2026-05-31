package services_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

type mockUserOrdersRepository struct {
	getUserOrdersFunc func(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error)
	completeUserOrder func(ctx context.Context, userID uuid.UUID, orderID uint) error
}

func (m *mockUserOrdersRepository) GetUserOrders(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
	return m.getUserOrdersFunc(ctx, id)
}

func (m *mockUserOrdersRepository) CompleteUserOrder(ctx context.Context, userID uuid.UUID, orderID uint) error {
	return m.completeUserOrder(ctx, userID, orderID)
}

func TestUserOrdersService_CompleteUserOrder(t *testing.T) {
	tests := []struct {
		name         string
		userID       uuid.UUID
		orderID      uint
		mockBehavior func(ctx context.Context, userID uuid.UUID, orderID uint) error
		wantErr      bool
		expectedErr  string
	}{
		{
			name:    "Success: Service calls repo correctly",
			userID:  uuid.New(),
			orderID: 1,
			mockBehavior: func(ctx context.Context, userID uuid.UUID, orderID uint) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "Error: Repo returns error",
			userID:  uuid.New(),
			orderID: 1,
			mockBehavior: func(ctx context.Context, userID uuid.UUID, orderID uint) error {
				return errors.New("insufficient energy")
			},
			wantErr:     true,
			expectedErr: "insufficient energy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockUserOrdersRepository{completeUserOrder: tt.mockBehavior}
			s := services.NewUserOrdersService(mRepo)

			err := s.CompleteUserOrder(context.Background(), tt.userID, tt.orderID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CompleteUserOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.expectedErr {
				t.Errorf("CompleteUserOrder() error = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

func TestUserOrdersService_GetUserOrders(t *testing.T) {
	uID := uuid.New()

	mockOrders := []domain.UserOrder{
		{ID: 1, UserID: uID, Status: "pending"},
		{ID: 2, UserID: uID, Status: "pending"},
	}
	tests := []struct {
		name    string // description of this test case
		repo    services.UserOrdersRepository
		id      uuid.UUID
		want    []domain.UserOrder
		wantErr bool
	}{
		{
			name: "Success: Returns orders from repository",
			id:   uID,
			repo: &mockUserOrdersRepository{
				getUserOrdersFunc: func(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
					return mockOrders, nil
				},
			},
			want:    mockOrders,
			wantErr: false,
		},
		{
			name: "Error: Repository fails",
			id:   uID,
			repo: &mockUserOrdersRepository{
				getUserOrdersFunc: func(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
					return nil, errors.New("database connection lost")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := services.NewUserOrdersService(tt.repo)
			got, gotErr := s.GetUserOrders(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUserOrders() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUserOrders() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserOrders() = %v, want %v", got, tt.want)
			}
		})
	}
}

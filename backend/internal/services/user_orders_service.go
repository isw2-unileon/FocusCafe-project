package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
)

// UserOrdersServiceInterface defines the methods that the UserOrdersService must implement
type UserOrdersServiceInterface interface {
	GetUserOrders(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error)
	CompleteUserOrder(ctx context.Context, userID uuid.UUID, orderID uint) error
}

// UserOrdersRepository defines the interface for user-related data operations
type UserOrdersRepository interface {
	GetUserOrders(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error)
	CompleteUserOrder(ctx context.Context, userID uuid.UUID, orderID uint) error
}

// UserOrdersService provides methods to handle user-related business logic
type UserOrdersService struct {
	repo UserOrdersRepository
}

// NewUserOrdersService creates a new instance of UserOrdersService with the given UserOrdersRepository
func NewUserOrdersService(repo UserOrdersRepository) *UserOrdersService {
	return &UserOrdersService{repo: repo}
}

// GetUserOrders retrieves the user orders information
func (s *UserOrdersService) GetUserOrders(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
	return s.repo.GetUserOrders(ctx, id)
}

// CompleteUserOrder completes the given user order
func (s *UserOrdersService) CompleteUserOrder(ctx context.Context, userID uuid.UUID, orderID uint) error {
	return s.repo.CompleteUserOrder(ctx, userID, orderID)
}

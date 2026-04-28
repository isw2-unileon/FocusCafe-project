package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

// UserRepository provides methods to interact with the database for user-related operations
type UserOrdersRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository with the given database connection
func NewUserOrdersRepository(db *gorm.DB) *UserOrdersRepository {
	return &UserOrdersRepository{db: db}
}

func (r *UserOrdersRepository) GetUserOrders(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
	var modelOrders []models.UserOrder

	// Query the database for the user_orders with the given ID
	err := r.db.WithContext(ctx).
		Preload("CafeOrder").
		Where("user_id = ? AND status = ?", id, "pending").
		Find(&modelOrders).Error
	if err != nil {
		return nil, err
	}

	// Convert modelUserOrders to domainUserOrders
	var domainUserOrders []domain.UserOrder
	for _, modelUserOrder := range modelOrders {
		domainUserOrders = append(domainUserOrders, domain.UserOrder{
			ID:          modelUserOrder.ID,
			UserID:      modelUserOrder.UserID,
			CafeOrderID: modelUserOrder.CafeOrderID,
			Status:      modelUserOrder.Status,
			CreatedAt:   modelUserOrder.CreatedAt,

			// Mapeamos también el café interno
			CafeOrder: &domain.CafeOrder{
				ID:            modelUserOrder.CafeOrder.ID,
				Name:          modelUserOrder.CafeOrder.Name,
				Description:   modelUserOrder.CafeOrder.Description,
				Category:      modelUserOrder.CafeOrder.Category,
				EnergyCost:    modelUserOrder.CafeOrder.EnergyCost,
				RewardXP:      modelUserOrder.CafeOrder.RewardXP,
				RequiredLevel: modelUserOrder.CafeOrder.RequiredLevel,
			},
		})
	}
	return domainUserOrders, nil
}

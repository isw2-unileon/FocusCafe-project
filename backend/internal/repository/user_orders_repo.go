package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

// UserOrdersRepository provides methods to interact with the database for user-related operations
type UserOrdersRepository struct {
	db *gorm.DB
}

// NewUserOrdersRepository creates a new instance of UserRepository with the given database connection
func NewUserOrdersRepository(db *gorm.DB) *UserOrdersRepository {
	return &UserOrdersRepository{db: db}
}

// GetUserOrders retrieves the user orders of a user
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

	if len(modelOrders) == 0 {
		fmt.Println("nada")
		if err := r.addCafeOrdersToUserByLevel(ctx, id); err != nil {
			return nil, err
		}

		r.db.WithContext(ctx).Preload("CafeOrder").Where("user_id = ? AND status = ?", id, "pending").Find(&modelOrders)
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

// addCafeOrdersToUserByLevel adds cafe orders to the user passed as an argument
func (r *UserOrdersRepository) addCafeOrdersToUserByLevel(ctx context.Context, userID uuid.UUID) error {
	var progress models.UserProgress
	if err := r.db.WithContext(ctx).Where(models.UserProgress{UserID: userID}).FirstOrCreate(&progress).Error; err != nil {
		return err
	}

	var availableCafes []models.CafeOrder

	// Search catalog for available cafe orders within the required level
	err := r.db.WithContext(ctx).
		Where("required_level <= ?", progress.Level).
		Order("RANDOM()").
		Limit(3).
		Find(&availableCafes).Error
	if err != nil {
		return err
	}
	// Create user orders and vinculate to the user
	for _, cafe := range availableCafes {
		newOrder := models.UserOrder{
			UserID:      userID,
			CafeOrderID: cafe.ID,
			Status:      "pending",
		}

		if err := r.db.WithContext(ctx).Create(&newOrder).Error; err != nil {
			return err
		}
	}

	return nil
}

// CompleteUserOrder completes the user order for the given user and cafe order
func (r *UserOrdersRepository) CompleteUserOrder(ctx context.Context, userID uuid.UUID, orderID uint) error {
	fmt.Printf("hola")
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Obtain order data and cafe order
		var userOrder models.UserOrder
		if err := tx.Preload("CafeOrder").First(&userOrder, orderID).Error; err != nil {
			return err
		}

		// 2. User progress
		var progress models.UserProgress
		if err := tx.Where("user_id = ?", userID).First(&progress).Error; err != nil {
			return err
		}

		// Validate energy
		if int64(progress.Energy) < userOrder.CafeOrder.EnergyCost {
			return errors.New("insufficient energy")
		}

		// Update user order status
		if err := tx.Model(&userOrder).Select("status").Updates(map[string]interface{}{"status": "completed"}).Error; err != nil {
			return err
		}

		// Update progress
		newXP := int64(progress.XP) + userOrder.CafeOrder.RewardXP
		newEnergy := int64(progress.Energy) - userOrder.CafeOrder.EnergyCost

		// Level logic
		newLevel := progress.Level
		if newXP >= (int64(progress.Level) * 100) {
			newLevel++
		}

		updates := map[string]interface{}{
			"XP":     newXP,
			"energy": newEnergy,
			"Level":  newLevel,
		}

		return tx.Model(&progress).Updates(updates).Error
	})
}

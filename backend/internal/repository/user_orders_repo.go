package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserOrdersRepository provides methods to interact with the database for user-related operations
type UserOrdersRepository struct {
	db *gorm.DB
}

// NewUserOrdersRepository creates a new instance of UserRepository with the given database connection
func NewUserOrdersRepository(db *gorm.DB) *UserOrdersRepository {
	return &UserOrdersRepository{db: db}
}

// GetUserOrders retrieves the user orders of a user, including personal and group orders.
func (r *UserOrdersRepository) GetUserOrders(ctx context.Context, id uuid.UUID) ([]domain.UserOrder, error) {
	// Obtain user's group_id
	var user models.User
	var groupID *int64
	if err := r.db.WithContext(ctx).Select("group_id").Where("id = ?", id).First(&user).Error; err == nil {
		groupID = user.GroupID
	}

	// Check personal orders and regenerate if empty
	var personalCount int64
	if err := r.db.WithContext(ctx).Model(&models.UserOrder{}).
		Where("user_id = ? AND group_id IS NULL AND status = ?", id, "pending").
		Count(&personalCount).Error; err != nil {
		return nil, err
	}
	if personalCount == 0 {
		if err := r.addCafeOrdersToUserByLevel(ctx, id); err != nil {
			return nil, err
		}
	}

	// Check group orders and regenerate if empty (only if user belongs to a group)
	if groupID != nil {
		var groupCount int64
		if err := r.db.WithContext(ctx).Model(&models.UserOrder{}).
			Where("group_id = ? AND status = ?", *groupID, "pending").
			Count(&groupCount).Error; err != nil {
			return nil, err
		}
		if groupCount == 0 {
			if err := r.addGroupOrders(ctx, *groupID); err != nil {
				return nil, err
			}
		}
	}

	// Final fetch: personal + group orders
	var modelOrders []models.UserOrder
	db := r.db.WithContext(ctx).Preload("CafeOrder").Where("status = ?", "pending")
	if groupID != nil {
		db = db.Where("(user_id = ? AND group_id IS NULL) OR group_id = ?", id, *groupID)
	} else {
		db = db.Where("user_id = ? AND group_id IS NULL", id)
	}
	if err := db.Find(&modelOrders).Error; err != nil {
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
			GroupID:     modelUserOrder.GroupID,
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

// addGroupOrders generates shared group orders based on the highest level in the group.
func (r *UserOrdersRepository) addGroupOrders(ctx context.Context, groupID int64) error {
	// Get the highest level among group members (default to 1 if no progress found)
	var maxLevel int
	row := r.db.WithContext(ctx).Raw(
		"SELECT COALESCE(MAX(p.level), 1) FROM user_progress p JOIN users u ON p.user_id = u.id WHERE u.group_id = ?",
		groupID,
	).Row()
	if err := row.Scan(&maxLevel); err != nil {
		return err
	}

	// Find available cafe orders
	var availableCafes []models.CafeOrder
	if err := r.db.WithContext(ctx).
		Where("required_level <= ?", maxLevel).
		Order("RANDOM()").
		Limit(3).
		Find(&availableCafes).Error; err != nil {
		return err
	}

	// Use the group leader as the user_id placeholder for group orders
	var group models.Group
	if err := r.db.WithContext(ctx).Select("leader_id").Where("id = ?", groupID).First(&group).Error; err != nil {
		return err
	}

	for _, cafe := range availableCafes {
		newOrder := models.UserOrder{
			UserID:      group.LeaderID,
			CafeOrderID: cafe.ID,
			Status:      "pending",
			GroupID:     &groupID,
		}
		if err := r.db.WithContext(ctx).Create(&newOrder).Error; err != nil {
			return err
		}
	}

	return nil
}

// CompleteUserOrder completes the user order for the given user and cafe order.
// If the order is a group order, XP is divided proportionally among all members,
// but only the completer pays the energy cost.
func (r *UserOrdersRepository) CompleteUserOrder(ctx context.Context, userID uuid.UUID, orderID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Obtain order data and cafe order
		var userOrder models.UserOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("CafeOrder").First(&userOrder, orderID).Error; err != nil {
			return err
		}

		if userOrder.Status == "completed" {
			return errors.New("order already completed")
		}

		// 2. Handle group order vs personal order
		if userOrder.GroupID != nil {
			return r.completeGroupOrder(tx, userID, userOrder)
		}

		return r.completePersonalOrder(tx, userID, userOrder)
	})
}

// completePersonalOrder handles completion of a personal user order.
func (r *UserOrdersRepository) completePersonalOrder(tx *gorm.DB, userID uuid.UUID, userOrder models.UserOrder) error {
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
}

// completeGroupOrder handles completion of a shared group order.
// Energy is paid by the completer; XP is divided equally among all group members.
func (r *UserOrdersRepository) completeGroupOrder(tx *gorm.DB, completerID uuid.UUID, userOrder models.UserOrder) error {
	if userOrder.GroupID == nil {
		return errors.New("group order missing group_id")
	}

	// 1. Get all group members
	var members []models.User
	if err := tx.Where("group_id = ?", *userOrder.GroupID).Find(&members).Error; err != nil {
		return err
	}
	if len(members) == 0 {
		return errors.New("group has no members")
	}

	// 2. Validate energy of the completer
	var completerProgress models.UserProgress
	if err := tx.Where("user_id = ?", completerID).First(&completerProgress).Error; err != nil {
		return err
	}
	if int64(completerProgress.Energy) < userOrder.CafeOrder.EnergyCost {
		return errors.New("insufficient energy")
	}

	// 3. Mark order as completed
	if err := tx.Model(&userOrder).Select("status").Updates(map[string]interface{}{"status": "completed"}).Error; err != nil {
		return err
	}

	// 4. Calculate XP per member
	memberCount := int64(len(members))
	xpPerMember := userOrder.CafeOrder.RewardXP / memberCount
	remainder := userOrder.CafeOrder.RewardXP % memberCount

	// 5. Update each member's progress
	for _, member := range members {
		var progress models.UserProgress
		if err := tx.Where("user_id = ?", member.ID).First(&progress).Error; err != nil {
			return err
		}

		xpToAdd := xpPerMember
		if member.ID == completerID {
			xpToAdd += remainder // completer gets the rounding remainder
		}

		newXP := int64(progress.XP) + xpToAdd
		newLevel := progress.Level
		if newXP >= (int64(progress.Level) * 100) {
			newLevel++
		}

		energyCost := int64(0)
		if member.ID == completerID {
			energyCost = userOrder.CafeOrder.EnergyCost
		}

		updates := map[string]interface{}{
			"XP":     newXP,
			"energy": int64(progress.Energy) - energyCost,
			"Level":  newLevel,
		}
		if err := tx.Model(&progress).Updates(updates).Error; err != nil {
			return err
		}
	}

	// 6. Regenerate group orders if none remain
	var remainingGroupOrders int64
	if err := tx.Model(&models.UserOrder{}).
		Where("group_id = ? AND status = ?", *userOrder.GroupID, "pending").
		Count(&remainingGroupOrders).Error; err != nil {
		return err
	}
	if remainingGroupOrders == 0 {
		return r.addGroupOrdersTx(tx, *userOrder.GroupID)
	}

	return nil
}

// addGroupOrdersTx is the transactional version of addGroupOrders for use inside a transaction.
func (r *UserOrdersRepository) addGroupOrdersTx(tx *gorm.DB, groupID int64) error {
	var maxLevel int
	row := tx.Raw(
		"SELECT COALESCE(MAX(p.level), 1) FROM user_progress p JOIN users u ON p.user_id = u.id WHERE u.group_id = ?",
		groupID,
	).Row()
	if err := row.Scan(&maxLevel); err != nil {
		return err
	}

	var availableCafes []models.CafeOrder
	if err := tx.Where("required_level <= ?", maxLevel).Order("RANDOM()").Limit(3).Find(&availableCafes).Error; err != nil {
		return err
	}

	var group models.Group
	if err := tx.Select("leader_id").Where("id = ?", groupID).First(&group).Error; err != nil {
		return err
	}

	for _, cafe := range availableCafes {
		newOrder := models.UserOrder{
			UserID:      group.LeaderID,
			CafeOrderID: cafe.ID,
			Status:      "pending",
			GroupID:     &groupID,
		}
		if err := tx.Create(&newOrder).Error; err != nil {
			return err
		}
	}

	return nil
}

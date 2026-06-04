package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

// GroupRepository provides methods to interact with the database for group-related operations.
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository creates a new instance of GroupRepository with the given database connection.
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// CreateGroup inserts a new group into the database.
func (r *GroupRepository) CreateGroup(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroupByInviteCode retrieves a group by its invite code.
func (r *GroupRepository) GetGroupByInviteCode(ctx context.Context, inviteCode string) (*models.Group, error) {
	var group models.Group
	err := r.db.WithContext(ctx).Where("invite_code = ?", inviteCode).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// GetGroupByID retrieves a group by its ID.
func (r *GroupRepository) GetGroupByID(ctx context.Context, id int64) (*models.Group, error) {
	var group models.Group
	err := r.db.WithContext(ctx).First(&group, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// AddUserToGroup updates the user's group_id to join a group.
func (r *GroupRepository) AddUserToGroup(ctx context.Context, userID uuid.UUID, groupID int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("group_id", groupID).Error
}

// GetUserGroupID retrieves the group_id of a user.
func (r *GroupRepository) GetUserGroupID(ctx context.Context, userID uuid.UUID) (*int64, error) {
	var user models.User
	err := r.db.WithContext(ctx).Select("group_id").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user.GroupID, nil
}

// IsUserInGroup checks if a user is already in any group.
func (r *GroupRepository) IsUserInGroup(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ? AND group_id IS NOT NULL", userID).Count(&count).Error
	return count > 0, err
}

// GetGroupMembers retrieves all users belonging to a group.
func (r *GroupRepository) GetGroupMembers(ctx context.Context, groupID int64) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&users).Error
	return users, err
}

// GetAllGroups retrieves all groups with their members.
func (r *GroupRepository) GetAllGroups(ctx context.Context) ([]models.Group, error) {
	var groups []models.Group
	err := r.db.WithContext(ctx).Preload("Users").Preload("Users.Progress").Find(&groups).Error
	return groups, err
}

// DeleteGroup removes a group from the database.
func (r *GroupRepository) DeleteGroup(ctx context.Context, groupID int64) error {
	return r.db.WithContext(ctx).Delete(&models.Group{}, groupID).Error
}

// RemoveAllUsersFromGroup sets group_id to NULL for all users in a group.
func (r *GroupRepository) RemoveAllUsersFromGroup(ctx context.Context, groupID int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("group_id = ?", groupID).Update("group_id", nil).Error
}

// IsGroupLeader checks if the given user is the leader of the group.
func (r *GroupRepository) IsGroupLeader(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Group{}).Where("id = ? AND leader_id = ?", groupID, userID).Count(&count).Error
	return count > 0, err
}

// RemoveUserFromGroup sets a single user's group_id to NULL.
func (r *GroupRepository) RemoveUserFromGroup(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("group_id", nil).Error
}

// DeleteGroupOrders removes all orders associated with a group.
func (r *GroupRepository) DeleteGroupOrders(ctx context.Context, groupID int64) error {
	return r.db.WithContext(ctx).Where("group_id = ?", groupID).Delete(&models.UserOrder{}).Error
}

// DeleteGroupWithDependencies removes a group and all its associated data atomically within a transaction.
func (r *GroupRepository) DeleteGroupWithDependencies(ctx context.Context, groupID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("group_id = ?", groupID).Update("group_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", groupID).Delete(&models.UserOrder{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Group{}, groupID).Error; err != nil {
			return err
		}
		return nil
	})
}

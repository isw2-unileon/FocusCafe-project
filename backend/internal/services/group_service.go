package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

// GroupRepository defines the interface for group-related data operations.
type GroupRepository interface {
	CreateGroup(ctx context.Context, group *models.Group) error
	GetGroupByInviteCode(ctx context.Context, inviteCode string) (*models.Group, error)
	GetGroupByID(ctx context.Context, id int64) (*models.Group, error)
	AddUserToGroup(ctx context.Context, userID uuid.UUID, groupID int64) error
	GetUserGroupID(ctx context.Context, userID uuid.UUID) (*int64, error)
	IsUserInGroup(ctx context.Context, userID uuid.UUID) (bool, error)
	GetAllGroups(ctx context.Context) ([]models.Group, error)
	DeleteGroup(ctx context.Context, groupID int64) error
	RemoveAllUsersFromGroup(ctx context.Context, groupID int64) error
	IsGroupLeader(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error)
	RemoveUserFromGroup(ctx context.Context, userID uuid.UUID) error
	DeleteGroupOrders(ctx context.Context, groupID int64) error
}

// GroupServiceInterface defines the methods that the GroupService must implement.
type GroupServiceInterface interface {
	CreateGroup(ctx context.Context, name string, leaderID uuid.UUID) (*domain.Group, error)
	JoinGroup(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.Group, error)
	GetAllGroups(ctx context.Context) ([]domain.GroupDetail, error)
	DeleteGroup(ctx context.Context, groupID int64) error
	LeaveGroup(ctx context.Context, userID uuid.UUID) error
	GetUserGroup(ctx context.Context, userID uuid.UUID) (*domain.Group, error)
}

// GroupService provides methods to handle group-related business logic.
type GroupService struct {
	repo GroupRepository
}

// NewGroupService creates a new instance of GroupService with the given GroupRepository.
func NewGroupService(repo GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

// generateInviteCode creates a random 6-character alphanumeric invite code.
func generateInviteCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// CreateGroup creates a new group with the given name and leader.
func (s *GroupService) CreateGroup(ctx context.Context, name string, leaderID uuid.UUID) (*domain.Group, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}

	// Check if user is already in a group
	inGroup, err := s.repo.IsUserInGroup(ctx, leaderID)
	if err != nil {
		return nil, err
	}
	if inGroup {
		return nil, errors.New("user is already in a group")
	}

	// Generate a unique invite code
	inviteCode := generateInviteCode()
	for {
		existing, err := s.repo.GetGroupByInviteCode(ctx, inviteCode)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			break
		}
		inviteCode = generateInviteCode()
	}

	group := &models.Group{
		Name:       name,
		InviteCode: inviteCode,
		LeaderID:   leaderID,
	}

	if err := s.repo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}

	// Automatically add the leader as a member of the group
	if err := s.repo.AddUserToGroup(ctx, leaderID, group.ID); err != nil {
		return nil, err
	}

	return &domain.Group{
		ID:         group.ID,
		Name:       group.Name,
		InviteCode: group.InviteCode,
		LeaderID:   group.LeaderID,
		CreatedAt:  group.CreatedAt,
	}, nil
}

// JoinGroup allows a user to join a group using an invite code.
func (s *GroupService) JoinGroup(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.Group, error) {
	if inviteCode == "" {
		return nil, errors.New("invite code is required")
	}

	// Check if user is already in a group
	inGroup, err := s.repo.IsUserInGroup(ctx, userID)
	if err != nil {
		return nil, err
	}
	if inGroup {
		return nil, errors.New("user is already in a group")
	}

	group, err := s.repo.GetGroupByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("invalid invite code: %s", inviteCode)
	}

	if err := s.repo.AddUserToGroup(ctx, userID, group.ID); err != nil {
		return nil, err
	}

	return &domain.Group{
		ID:         group.ID,
		Name:       group.Name,
		InviteCode: group.InviteCode,
		LeaderID:   group.LeaderID,
		CreatedAt:  group.CreatedAt,
	}, nil
}

// GetAllGroups retrieves all groups with their members.
func (s *GroupService) GetAllGroups(ctx context.Context) ([]domain.GroupDetail, error) {
	groups, err := s.repo.GetAllGroups(ctx)
	if err != nil {
		return nil, err
	}

	var details []domain.GroupDetail
	for _, g := range groups {
		detail := domain.GroupDetail{
			ID:         g.ID,
			Name:       g.Name,
			InviteCode: g.InviteCode,
			LeaderID:   g.LeaderID,
			CreatedAt:  g.CreatedAt,
			Members:    make([]domain.GroupMember, 0, len(g.Users)),
		}

		for _, u := range g.Users {
			member := domain.GroupMember{
				ID:        u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Email:     u.Email,
				Level:     1,
			}
			if u.Progress != nil {
				member.Level = u.Progress.Level
			}
			detail.Members = append(detail.Members, member)
		}

		details = append(details, detail)
	}

	return details, nil
}

// DeleteGroup removes a group and all its associated data.
func (s *GroupService) DeleteGroup(ctx context.Context, groupID int64) error {
	// 1. Remove all users from the group
	if err := s.repo.RemoveAllUsersFromGroup(ctx, groupID); err != nil {
		return err
	}

	// 2. Delete group orders
	if err := s.repo.DeleteGroupOrders(ctx, groupID); err != nil {
		return err
	}

	// 3. Delete the group itself
	return s.repo.DeleteGroup(ctx, groupID)
}

// LeaveGroup allows a user to leave their current group.
// Leaders must use DeleteGroup instead.
func (s *GroupService) LeaveGroup(ctx context.Context, userID uuid.UUID) error {
	// Get user's current group
	groupID, err := s.repo.GetUserGroupID(ctx, userID)
	if err != nil {
		return err
	}
	if groupID == nil {
		return errors.New("user is not in any group")
	}

	// Check if user is the leader
	isLeader, err := s.repo.IsGroupLeader(ctx, userID, *groupID)
	if err != nil {
		return err
	}
	if isLeader {
		return errors.New("group leader cannot leave the group, use delete instead")
	}

	// Remove user from group
	return s.repo.RemoveUserFromGroup(ctx, userID)
}

// GetUserGroup retrieves the group the user currently belongs to.
func (s *GroupService) GetUserGroup(ctx context.Context, userID uuid.UUID) (*domain.Group, error) {
	groupID, err := s.repo.GetUserGroupID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if groupID == nil {
		return nil, errors.New("user is not in any group")
	}

	group, err := s.repo.GetGroupByID(ctx, *groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("group not found")
	}

	return &domain.Group{
		ID:         group.ID,
		Name:       group.Name,
		InviteCode: group.InviteCode,
		LeaderID:   group.LeaderID,
		CreatedAt:  group.CreatedAt,
	}, nil
}

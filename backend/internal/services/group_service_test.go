package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

// mockGroupRepository is a test double for the GroupRepository.
type mockGroupRepository struct {
	createGroupFunc                 func(ctx context.Context, group *models.Group) error
	getGroupByInviteCodeFunc        func(ctx context.Context, inviteCode string) (*models.Group, error)
	getGroupByIDFunc                func(ctx context.Context, id int64) (*models.Group, error)
	addUserToGroupFunc              func(ctx context.Context, userID uuid.UUID, groupID int64) error
	getUserGroupIDFunc              func(ctx context.Context, userID uuid.UUID) (*int64, error)
	isUserInGroupFunc               func(ctx context.Context, userID uuid.UUID) (bool, error)
	getAllGroupsFunc                func(ctx context.Context) ([]models.Group, error)
	deleteGroupFunc                 func(ctx context.Context, groupID int64) error
	removeAllUsersFromGroup         func(ctx context.Context, groupID int64) error
	isGroupLeaderFunc               func(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error)
	removeUserFromGroupFunc         func(ctx context.Context, userID uuid.UUID) error
	deleteGroupOrdersFunc           func(ctx context.Context, groupID int64) error
	deleteGroupWithDependenciesFunc func(ctx context.Context, groupID int64) error
}

func (m *mockGroupRepository) CreateGroup(ctx context.Context, group *models.Group) error {
	if m.createGroupFunc != nil {
		return m.createGroupFunc(ctx, group)
	}
	return errors.New("createGroup not mocked")
}

func (m *mockGroupRepository) GetGroupByInviteCode(ctx context.Context, inviteCode string) (*models.Group, error) {
	if m.getGroupByInviteCodeFunc != nil {
		return m.getGroupByInviteCodeFunc(ctx, inviteCode)
	}
	return nil, errors.New("getGroupByInviteCode not mocked")
}

func (m *mockGroupRepository) GetGroupByID(ctx context.Context, id int64) (*models.Group, error) {
	if m.getGroupByIDFunc != nil {
		return m.getGroupByIDFunc(ctx, id)
	}
	return nil, errors.New("getGroupByID not mocked")
}

func (m *mockGroupRepository) AddUserToGroup(ctx context.Context, userID uuid.UUID, groupID int64) error {
	if m.addUserToGroupFunc != nil {
		return m.addUserToGroupFunc(ctx, userID, groupID)
	}
	return errors.New("addUserToGroup not mocked")
}

func (m *mockGroupRepository) GetUserGroupID(ctx context.Context, userID uuid.UUID) (*int64, error) {
	if m.getUserGroupIDFunc != nil {
		return m.getUserGroupIDFunc(ctx, userID)
	}
	return nil, errors.New("getUserGroupID not mocked")
}

func (m *mockGroupRepository) IsUserInGroup(ctx context.Context, userID uuid.UUID) (bool, error) {
	if m.isUserInGroupFunc != nil {
		return m.isUserInGroupFunc(ctx, userID)
	}
	return false, errors.New("isUserInGroup not mocked")
}

func (m *mockGroupRepository) GetAllGroups(ctx context.Context) ([]models.Group, error) {
	if m.getAllGroupsFunc != nil {
		return m.getAllGroupsFunc(ctx)
	}
	return nil, errors.New("getAllGroups not mocked")
}

func (m *mockGroupRepository) DeleteGroup(ctx context.Context, groupID int64) error {
	if m.deleteGroupFunc != nil {
		return m.deleteGroupFunc(ctx, groupID)
	}
	return errors.New("deleteGroup not mocked")
}

func (m *mockGroupRepository) RemoveAllUsersFromGroup(ctx context.Context, groupID int64) error {
	if m.removeAllUsersFromGroup != nil {
		return m.removeAllUsersFromGroup(ctx, groupID)
	}
	return errors.New("removeAllUsersFromGroup not mocked")
}

func (m *mockGroupRepository) IsGroupLeader(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error) {
	if m.isGroupLeaderFunc != nil {
		return m.isGroupLeaderFunc(ctx, userID, groupID)
	}
	return false, errors.New("isGroupLeader not mocked")
}

func (m *mockGroupRepository) RemoveUserFromGroup(ctx context.Context, userID uuid.UUID) error {
	if m.removeUserFromGroupFunc != nil {
		return m.removeUserFromGroupFunc(ctx, userID)
	}
	return errors.New("removeUserFromGroup not mocked")
}

func (m *mockGroupRepository) DeleteGroupOrders(ctx context.Context, groupID int64) error {
	if m.deleteGroupOrdersFunc != nil {
		return m.deleteGroupOrdersFunc(ctx, groupID)
	}
	return errors.New("deleteGroupOrders not mocked")
}

func (m *mockGroupRepository) DeleteGroupWithDependencies(ctx context.Context, groupID int64) error {
	if m.deleteGroupWithDependenciesFunc != nil {
		return m.deleteGroupWithDependenciesFunc(ctx, groupID)
	}
	return errors.New("deleteGroupWithDependencies not mocked")
}

// ============================================
// TestGroupService_CreateGroup
// ============================================

type createGroupTestCase struct {
	name           string
	groupName      string
	leaderID       uuid.UUID
	isInGroup      bool
	isInGroupErr   error
	codeCollides   bool
	getByCodeErr   error
	createErr      error
	addUserErr     error
	wantErr        bool
	expectedErr    string
	expectedLength int
}

func TestGroupService_CreateGroup(t *testing.T) {
	leaderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []createGroupTestCase{
		{
			name:           "Success: Creates group with unique code",
			groupName:      "The A-Team",
			leaderID:       leaderID,
			isInGroup:      false,
			codeCollides:   false,
			wantErr:        false,
			expectedLength: 6,
		},
		{
			name:           "Success: Re-generates code if collision occurs",
			groupName:      "The Collision Team",
			leaderID:       leaderID,
			isInGroup:      false,
			codeCollides:   true,
			wantErr:        false,
			expectedLength: 6,
		},
		{
			name:        "Error: User already in a group",
			groupName:   "The B-Team",
			leaderID:    leaderID,
			isInGroup:   true,
			wantErr:     true,
			expectedErr: "user is already in a group",
		},
		{
			name:        "Error: Empty group name",
			groupName:   "",
			leaderID:    leaderID,
			wantErr:     true,
			expectedErr: "group name is required",
		},
		{
			name:         "Error: IsUserInGroup database error",
			groupName:    "Fail Group",
			leaderID:     leaderID,
			isInGroupErr: errors.New("db connection failure"),
			wantErr:      true,
			expectedErr:  "db connection failure",
		},
		{
			name:         "Error: GetGroupByInviteCode database error",
			groupName:    "Fail Code Group",
			leaderID:     leaderID,
			getByCodeErr: errors.New("lookup error"),
			wantErr:      true,
			expectedErr:  "lookup error",
		},
		{
			name:        "Error: CreateGroup database error",
			groupName:   "Fail Create Group",
			leaderID:    leaderID,
			createErr:   errors.New("insert error"),
			wantErr:     true,
			expectedErr: "insert error",
		},
		{
			name:        "Error: AddUserToGroup database error",
			groupName:   "Fail Add Leader Group",
			leaderID:    leaderID,
			addUserErr:  errors.New("binding error"),
			wantErr:     true,
			expectedErr: "binding error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCreateGroupSubtest(t, tt)
		})
	}
}

func runCreateGroupSubtest(t *testing.T, tt createGroupTestCase) {
	collisionCalled := false

	mRepo := &mockGroupRepository{
		isUserInGroupFunc: func(ctx context.Context, userID uuid.UUID) (bool, error) {
			return tt.isInGroup, tt.isInGroupErr
		},
		getGroupByInviteCodeFunc: func(ctx context.Context, code string) (*models.Group, error) {
			if tt.getByCodeErr != nil {
				return nil, tt.getByCodeErr
			}
			if tt.codeCollides && !collisionCalled {
				collisionCalled = true
				return &models.Group{ID: 99, InviteCode: code}, nil
			}
			return nil, nil
		},
		createGroupFunc: func(ctx context.Context, group *models.Group) error {
			return tt.createErr
		},
		addUserToGroupFunc: func(ctx context.Context, userID uuid.UUID, groupID int64) error {
			return tt.addUserErr
		},
	}

	s := services.NewGroupService(mRepo)
	group, err := s.CreateGroup(context.Background(), tt.groupName, tt.leaderID)

	if (err != nil) != tt.wantErr {
		t.Fatalf("CreateGroup() error = %v, wantErr %v", err, tt.wantErr)
	}

	if tt.wantErr {
		if err.Error() != tt.expectedErr {
			t.Errorf("CreateGroup() error = %v, want %v", err, tt.expectedErr)
		}
		return
	}

	if group == nil {
		t.Fatal("CreateGroup() returned nil group")
	}

	if len(group.InviteCode) != tt.expectedLength {
		t.Errorf("CreateGroup() invite code length = %v, want %v", len(group.InviteCode), tt.expectedLength)
	}
}

// ============================================
// TestGroupService_JoinGroup
// ============================================

func TestGroupService_JoinGroup(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	tests := []struct {
		name          string
		inviteCode    string
		mockIsInGroup func(ctx context.Context, userID uuid.UUID) (bool, error)
		mockGetByCode func(ctx context.Context, code string) (*models.Group, error)
		mockAddUser   func(ctx context.Context, userID uuid.UUID, groupID int64) error
		wantErr       bool
		expectedErr   string
	}{
		{
			name:       "Success: Joins group with valid code",
			inviteCode: "AB12CD",
			mockIsInGroup: func(ctx context.Context, userID uuid.UUID) (bool, error) {
				return false, nil
			},
			mockGetByCode: func(ctx context.Context, code string) (*models.Group, error) {
				return &models.Group{ID: 1, Name: "The A-Team", InviteCode: code, LeaderID: uuid.New()}, nil
			},
			mockAddUser: func(ctx context.Context, userID uuid.UUID, groupID int64) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:       "Error: IsUserInGroup failure",
			inviteCode: "AB12CD",
			mockIsInGroup: func(ctx context.Context, userID uuid.UUID) (bool, error) {
				return false, errors.New("network failure")
			},
			wantErr:     true,
			expectedErr: "network failure",
		},
		{
			name:       "Error: User already in group",
			inviteCode: "AB12CD",
			mockIsInGroup: func(ctx context.Context, userID uuid.UUID) (bool, error) {
				return true, nil
			},
			wantErr:     true,
			expectedErr: "user is already in a group",
		},
		{
			name:       "Error: GetGroupByInviteCode database error",
			inviteCode: "AB12CD",
			mockIsInGroup: func(ctx context.Context, userID uuid.UUID) (bool, error) {
				return false, nil
			},
			mockGetByCode: func(ctx context.Context, code string) (*models.Group, error) {
				return nil, errors.New("query failed")
			},
			wantErr:     true,
			expectedErr: "query failed",
		},
		{
			name:       "Error: Invalid invite code",
			inviteCode: "ZZZZZZ",
			mockIsInGroup: func(ctx context.Context, userID uuid.UUID) (bool, error) {
				return false, nil
			},
			mockGetByCode: func(ctx context.Context, code string) (*models.Group, error) {
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "invalid invite code: ZZZZZZ",
		},
		{
			name:       "Error: Empty invite code",
			inviteCode: "",
			mockIsInGroup: func(ctx context.Context, userID uuid.UUID) (bool, error) {
				return false, nil
			},
			wantErr:     true,
			expectedErr: "invite code is required",
		},
		{
			name:       "Error: AddUserToGroup failure",
			inviteCode: "AB12CD",
			mockIsInGroup: func(ctx context.Context, userID uuid.UUID) (bool, error) {
				return false, nil
			},
			mockGetByCode: func(ctx context.Context, code string) (*models.Group, error) {
				return &models.Group{ID: 1, Name: "The A-Team", InviteCode: code, LeaderID: uuid.New()}, nil
			},
			mockAddUser: func(ctx context.Context, userID uuid.UUID, groupID int64) error {
				return errors.New("capacity exceeded")
			},
			wantErr:     true,
			expectedErr: "capacity exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockGroupRepository{
				isUserInGroupFunc:        tt.mockIsInGroup,
				getGroupByInviteCodeFunc: tt.mockGetByCode,
				addUserToGroupFunc:       tt.mockAddUser,
			}
			s := services.NewGroupService(mRepo)

			group, err := s.JoinGroup(context.Background(), tt.inviteCode, userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("JoinGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.expectedErr {
					t.Errorf("JoinGroup() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if group == nil {
				t.Error("JoinGroup() returned nil group")
			}
		})
	}
}

// ============================================
// TestGroupService_GetAllGroups
// ============================================

func TestGroupService_GetAllGroups(t *testing.T) {
	tests := []struct {
		name         string
		mockBehavior func(ctx context.Context) ([]models.Group, error)
		wantErr      bool
		expectedErr  string
		expectedLen  int
	}{
		{
			name: "Success: Returns groups with members",
			mockBehavior: func(ctx context.Context) ([]models.Group, error) {
				return []models.Group{
					{
						ID:         1,
						Name:       "The A-Team",
						InviteCode: "AB12CD",
						LeaderID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
						Users: []models.User{
							{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), FirstName: "Alice", Email: "alice@test.com"},
						},
					},
				}, nil
			},
			wantErr:     false,
			expectedLen: 1,
		},
		{
			name: "Success: Returns empty list",
			mockBehavior: func(ctx context.Context) ([]models.Group, error) {
				return []models.Group{}, nil
			},
			wantErr:     false,
			expectedLen: 0,
		},
		{
			name: "Error: Repository failure",
			mockBehavior: func(ctx context.Context) ([]models.Group, error) {
				return nil, errors.New("database error")
			},
			wantErr:     true,
			expectedErr: "database error",
		},
		{
			name: "Success: Returns groups with members and progress mapped",
			mockBehavior: func(ctx context.Context) ([]models.Group, error) {
				return []models.Group{
					{
						ID:         1,
						Name:       "The A-Team",
						InviteCode: "AB12CD",
						LeaderID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
						Users: []models.User{
							{
								ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
								FirstName: "Alice",
								Email:     "alice@test.com",
								Progress:  &models.UserProgress{Level: 5},
							},
							{
								ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440099"),
								FirstName: "Bob",
								Email:     "bob@test.com",
								Progress:  nil,
							},
						},
					},
				}, nil
			},
			wantErr:     false,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockGroupRepository{getAllGroupsFunc: tt.mockBehavior}
			s := services.NewGroupService(mRepo)

			groups, err := s.GetAllGroups(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.expectedErr {
					t.Errorf("GetAllGroups() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if len(groups) != tt.expectedLen {
				t.Errorf("GetAllGroups() len = %v, want %v", len(groups), tt.expectedLen)
			}
		})
	}
}

// ============================================
// TestGroupService_DeleteGroup
// ============================================

func TestGroupService_DeleteGroup(t *testing.T) {
	tests := []struct {
		name                            string
		groupID                         int64
		mockDeleteGroupWithDependencies func(ctx context.Context, groupID int64) error
		wantErr                         bool
		expectedErr                     string
	}{
		{
			name:    "Success: Deletes group and cleans up",
			groupID: 1,
			mockDeleteGroupWithDependencies: func(ctx context.Context, groupID int64) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "Error: Transaction fails",
			groupID: 1,
			mockDeleteGroupWithDependencies: func(ctx context.Context, groupID int64) error {
				return errors.New("foreign key constraint")
			},
			wantErr:     true,
			expectedErr: "foreign key constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockGroupRepository{
				deleteGroupWithDependenciesFunc: tt.mockDeleteGroupWithDependencies,
			}
			s := services.NewGroupService(mRepo)

			err := s.DeleteGroup(context.Background(), tt.groupID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.expectedErr {
				t.Errorf("DeleteGroup() error = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

// ============================================
// TestGroupService_LeaveGroup
// ============================================

func TestGroupService_LeaveGroup(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	groupID := int64(1)

	tests := []struct {
		name           string
		mockGetGroupID func(ctx context.Context, userID uuid.UUID) (*int64, error)
		mockIsLeader   func(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error)
		mockRemoveUser func(ctx context.Context, userID uuid.UUID) error
		wantErr        bool
		expectedErr    string
	}{
		{
			name: "Success: Member leaves group",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return &groupID, nil
			},
			mockIsLeader: func(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error) {
				return false, nil
			},
			mockRemoveUser: func(ctx context.Context, userID uuid.UUID) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "Error: GetUserGroupID repository failure",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return nil, errors.New("lookup failed")
			},
			wantErr:     true,
			expectedErr: "lookup failed",
		},
		{
			name: "Error: Leader cannot leave",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return &groupID, nil
			},
			mockIsLeader: func(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error) {
				return true, nil
			},
			wantErr:     true,
			expectedErr: "group leader cannot leave the group, use delete instead",
		},
		{
			name: "Error: IsGroupLeader repository failure",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return &groupID, nil
			},
			mockIsLeader: func(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error) {
				return false, errors.New("leader verification failed")
			},
			wantErr:     true,
			expectedErr: "leader verification failed",
		},
		{
			name: "Error: User not in any group",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "user is not in any group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockGroupRepository{
				getUserGroupIDFunc:      tt.mockGetGroupID,
				isGroupLeaderFunc:       tt.mockIsLeader,
				removeUserFromGroupFunc: tt.mockRemoveUser,
			}
			s := services.NewGroupService(mRepo)

			err := s.LeaveGroup(context.Background(), userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("LeaveGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.expectedErr {
				t.Errorf("LeaveGroup() error = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

// ============================================
// TestGroupService_GetUserGroup
// ============================================

func TestGroupService_GetUserGroup(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
	groupID := int64(1)

	tests := []struct {
		name           string
		mockGetGroupID func(ctx context.Context, userID uuid.UUID) (*int64, error)
		mockGetByID    func(ctx context.Context, id int64) (*models.Group, error)
		wantErr        bool
		expectedErr    string
		expectedName   string
	}{
		{
			name: "Success: Returns user's group",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return &groupID, nil
			},
			mockGetByID: func(ctx context.Context, id int64) (*models.Group, error) {
				return &models.Group{ID: id, Name: "The A-Team"}, nil
			},
			wantErr:      false,
			expectedName: "The A-Team",
		},
		{
			name: "Error: GetUserGroupID failure",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return nil, errors.New("db crash")
			},
			wantErr:     true,
			expectedErr: "db crash",
		},
		{
			name: "Error: User not in any group",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "user is not in any group",
		},
		{
			name: "Error: GetGroupByID failure",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return &groupID, nil
			},
			mockGetByID: func(ctx context.Context, id int64) (*models.Group, error) {
				return nil, errors.New("fetch error")
			},
			wantErr:     true,
			expectedErr: "fetch error",
		},
		{
			name: "Error: Group not found (nil record)",
			mockGetGroupID: func(ctx context.Context, userID uuid.UUID) (*int64, error) {
				return &groupID, nil
			},
			mockGetByID: func(ctx context.Context, id int64) (*models.Group, error) {
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "group not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockGroupRepository{
				getUserGroupIDFunc: tt.mockGetGroupID,
				getGroupByIDFunc:   tt.mockGetByID,
			}
			s := services.NewGroupService(mRepo)

			group, err := s.GetUserGroup(context.Background(), userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.expectedErr {
					t.Errorf("GetUserGroup() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if group.Name != tt.expectedName {
				t.Errorf("GetUserGroup() name = %v, want %v", group.Name, tt.expectedName)
			}
		})
	}
}

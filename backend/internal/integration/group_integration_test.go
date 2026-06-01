package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

func TestGroups_Integration(t *testing.T) {
	db, h := setupTestApp()
	u1, u2 := uuid.New(), uuid.New()

	// Seed Users to avoid "record not found" in repository
	db.Create(&models.User{ID: u1, Username: "leader", Email: "leader@test.com", FirstName: "L", LastName: "D"})
	db.Create(&models.User{ID: u2, Username: "member", Email: "member@test.com", FirstName: "M", LastName: "B"})

	var sharedInviteCode string

	// Main execution orchestration flow
	t.Run("Create Group", func(t *testing.T) {
		sharedInviteCode = testCreateGroup(t, db, h, u1)
	})

	t.Run("Join Group", func(t *testing.T) {
		testJoinGroup(t, db, h, u2, sharedInviteCode)
	})

	t.Run("Leave Group", func(t *testing.T) {
		testLeaveGroup(t, db, h, u2)
	})

	t.Run("Delete Group (Leader)", func(t *testing.T) {
		testDeleteGroup(t, db, h, u1, sharedInviteCode)
	})
}

func testCreateGroup(t *testing.T, db *gorm.DB, h *handlers.Handler, leaderID uuid.UUID) string {
	groupName := "Test Group"
	body, _ := json.Marshal(map[string]string{"name": groupName})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/groups", bytes.NewBuffer(body))

	c, _ := setupSubtestContext(w, req, leaderID)
	h.CreateGroup(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var createdGroup models.Group
	if err := json.Unmarshal(w.Body.Bytes(), &createdGroup); err != nil {
		t.Fatalf("failed to unmarshal group: %v", err)
	}

	if createdGroup.Name != groupName {
		t.Errorf("expected group name %s, got %s", groupName, createdGroup.Name)
	}

	// Verify DB
	var user models.User
	db.First(&user, leaderID)
	if user.GroupID == nil || *user.GroupID != createdGroup.ID {
		t.Errorf("expected user to be in group %d, got %v", createdGroup.ID, user.GroupID)
	}

	return createdGroup.InviteCode
}

func testJoinGroup(t *testing.T, db *gorm.DB, h *handlers.Handler, memberID uuid.UUID, inviteCode string) {
	var group models.Group
	if err := db.Where("invite_code = ?", inviteCode).First(&group).Error; err != nil {
		t.Fatalf("failed to find group in DB for joining: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"invite_code": group.InviteCode})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/groups/join", bytes.NewBuffer(body))

	c, _ := setupSubtestContext(w, req, memberID)
	h.JoinGroup(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify DB
	var user models.User
	db.First(&user, memberID)
	if user.GroupID == nil || *user.GroupID != group.ID {
		t.Errorf("expected user to be in group %d, got %v", group.ID, user.GroupID)
	}
}

func testLeaveGroup(t *testing.T, db *gorm.DB, h *handlers.Handler, memberID uuid.UUID) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/groups/leave", nil)

	c, _ := setupSubtestContext(w, req, memberID)
	h.LeaveGroup(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify DB
	var user models.User
	db.First(&user, memberID)
	if user.GroupID != nil {
		t.Errorf("expected user to have no group, got %d", *user.GroupID)
	}
}

func testDeleteGroup(t *testing.T, db *gorm.DB, h *handlers.Handler, leaderID uuid.UUID, inviteCode string) {
	var group models.Group
	if err := db.Where("invite_code = ?", inviteCode).First(&group).Error; err != nil {
		t.Fatalf("failed to find group before deletion: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/groups", nil)

	c, _ := setupSubtestContext(w, req, leaderID)
	h.DeleteGroup(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify DB: the groupID should not exist
	var count int64
	db.Model(&models.Group{}).Where("id = ?", group.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected group %d to be deleted, but it still exists", group.ID)
	}
}

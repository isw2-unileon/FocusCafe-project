package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

func TestGroups_Integration(t *testing.T) {
	_, db, h := setupTestApp()
	u1, u2 := uuid.New(), uuid.New()

	// Seed Users to avoid "record not found" in repository
	db.Create(&models.User{ID: u1, Username: "leader", Email: "leader@test.com", FirstName: "L", LastName: "D"})
	db.Create(&models.User{ID: u2, Username: "member", Email: "member@test.com", FirstName: "M", LastName: "B"})

	var sharedInviteCode string

	t.Run("Create Group", func(t *testing.T) {
		groupName := "Test Group"
		body, _ := json.Marshal(map[string]string{"name": groupName})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/groups", bytes.NewBuffer(body))

		c, _ := setupSubtestContext(w, req, u1)
		h.CreateGroup(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d. Body: %s", w.Code, w.Body.String()) // <--- Cambiado a Fatalf
		}

		var createdGroup models.Group
		if err := json.Unmarshal(w.Body.Bytes(), &createdGroup); err != nil {
			t.Fatalf("failed to unmarshal group: %v", err)
		}

		if createdGroup.Name != groupName {
			t.Errorf("expected group name %s, got %s", groupName, createdGroup.Name)
		}

		sharedInviteCode = createdGroup.InviteCode

		// Verify DB
		var user models.User
		db.First(&user, u1)
		if user.GroupID == nil || *user.GroupID != createdGroup.ID {
			t.Errorf("expected user to be in group %d, got %v", createdGroup.ID, user.GroupID)
		}
	})

	t.Run("Join Group", func(t *testing.T) {
		var group models.Group
		if err := db.Where("invite_code = ?", sharedInviteCode).First(&group).Error; err != nil {
			t.Fatalf("failed to find group in DB for joining: %v", err)
		}

		body, _ := json.Marshal(map[string]string{"invite_code": group.InviteCode})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/groups/join", bytes.NewBuffer(body))

		c, _ := setupSubtestContext(w, req, u2)
		h.JoinGroup(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		// Verify DB
		var user models.User
		db.First(&user, u2)
		if user.GroupID == nil || *user.GroupID != group.ID {
			t.Errorf("expected user to be in group %d, got %v", group.ID, user.GroupID)
		}
	})

	t.Run("Leave Group", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/groups/leave", nil)

		c, _ := setupSubtestContext(w, req, u2)
		h.LeaveGroup(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		// Verify DB
		var user models.User
		db.First(&user, u2)
		if user.GroupID != nil {
			t.Errorf("expected user to have no group, got %d", *user.GroupID)
		}
	})

	t.Run("Delete Group (Leader)", func(t *testing.T) {
		var group models.Group
		if err := db.Where("invite_code = ?", sharedInviteCode).First(&group).Error; err != nil {
			t.Fatalf("failed to find group before deletion: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/groups", nil)

		c, _ := setupSubtestContext(w, req, u1)
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
	})
}

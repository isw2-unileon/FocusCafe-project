package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateGroupRequest represents the request body for creating a group.
type CreateGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

// JoinGroupRequest represents the request body for joining a group.
type JoinGroupRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// CreateGroup handles the creation of a new study/cafe group.
func (h *Handler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group name is required"})
		return
	}

	uid, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	group, err := h.GroupService.CreateGroup(c.Request.Context(), req.Name, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// JoinGroup handles joining a group via invite code.
func (h *Handler) JoinGroup(c *gin.Context) {
	var req JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invite code is required"})
		return
	}

	uid, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	group, err := h.GroupService.JoinGroup(c.Request.Context(), req.InviteCode, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// LeaveGroup allows the current user to leave their group.
func (h *Handler) LeaveGroup(c *gin.Context) {
	uid, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.GroupService.LeaveGroup(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "left group successfully"})
}

// DeleteGroup allows the group leader to delete their group.
func (h *Handler) DeleteGroup(c *gin.Context) {
	uid, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get the user's current group
	group, err := h.GroupService.GetUserGroup(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the user is the leader
	if group.LeaderID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the group leader can delete the group"})
		return
	}

	if err := h.GroupService.DeleteGroup(c.Request.Context(), group.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group deleted successfully"})
}

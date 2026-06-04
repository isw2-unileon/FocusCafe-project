package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterRequest Contains the data required to register a new user.
type RegisterRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// Register is the main handler
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := validateRegisterRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.createAuthUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	if err := h.createUserProfile(userID, req, "user"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.createUserProgress(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         userID,
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	})
}

// validateRegisterRequest validates form data
func validateRegisterRequest(req *RegisterRequest) error {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)

	if req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("error: first name and surname are required")
	}
	if req.Email == "" {
		return fmt.Errorf("error: email is mandatory")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("error: the password must be at least 6 characters long")
	}
	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("error: the passwords do not match")
	}
	return nil
}

// createAuthUser creates the user in Supabase Auth and return its UUID
func (h *Handler) createAuthUser(email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req, _ := http.NewRequest(http.MethodPost,
		h.SupabaseURL+"/auth/v1/signup",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", h.SupabaseKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("connection error while creating user")
	}
	defer resp.Body.Close()

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("error processing auth service response")
	}

	if resp.StatusCode != http.StatusOK {
		if msg, ok := data["msg"].(string); ok && msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		if message, ok := data["message"].(string); ok && message != "" {
			return "", fmt.Errorf("%s", message)
		}
		if desc, ok := data["error_description"].(string); ok && desc != "" {
			return "", fmt.Errorf("%s", desc)
		}
		return "", fmt.Errorf("could not create user (status %d)", resp.StatusCode)
	}

	return extractUserID(data)
}

// extractUserID extracts the user's UUID from the Supabase response
func extractUserID(data map[string]any) (string, error) {
	userMap, ok := data["user"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("unexpected response format from auth service")
	}

	userID, ok := userMap["id"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in auth response")
	}

	return userID, nil
}

// createUserProfile inserts the user profile into the public.users table
func (h *Handler) createUserProfile(userID string, req RegisterRequest, role string) error {
	if role == "" {
		role = "user"
	}
	username := strings.Split(req.Email, "@")[0]

	body, _ := json.Marshal(map[string]string{
		"id":         userID,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"username":   username,
		"email":      req.Email,
		"role":       role,
	})

	profileReq, _ := http.NewRequest(
		http.MethodPost,
		h.SupabaseURL+"/rest/v1/users",
		bytes.NewBuffer(body),
	)
	profileReq.Header.Set("Content-Type", "application/json")
	profileReq.Header.Set("apikey", h.SupabaseServiceRoleKey)
	profileReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.SupabaseServiceRoleKey))
	profileReq.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(profileReq)
	if err != nil {
		return fmt.Errorf("connection error while saving user profile")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var profileErr map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&profileErr); err != nil {
			return fmt.Errorf("error saving profile (status %d)", resp.StatusCode)
		}

		if msg, ok := profileErr["message"].(string); ok && msg != "" {
			return fmt.Errorf("database error: %s", msg)
		}
		return fmt.Errorf("failed to save profile")
	}

	return nil
}

// createUserProgress inserts the user's initial progress into public.user_progress
func (h *Handler) createUserProgress(userID string) error {
	body, _ := json.Marshal(map[string]any{
		"user_id": userID,
		"energy":  500,
		"level":   1,
		"xp":      0,
	})

	progressReq, _ := http.NewRequest(
		http.MethodPost,
		h.SupabaseURL+"/rest/v1/user_progress",
		bytes.NewBuffer(body),
	)
	progressReq.Header.Set("Content-Type", "application/json")
	progressReq.Header.Set("apikey", h.SupabaseServiceRoleKey)
	progressReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.SupabaseServiceRoleKey))
	progressReq.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(progressReq)
	if err != nil {
		return fmt.Errorf("connection error while saving user progress")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var progressErr map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&progressErr); err != nil {
			return fmt.Errorf("error saving progress (status %d)", resp.StatusCode)
		}

		if msg, ok := progressErr["message"].(string); ok && msg != "" {
			return fmt.Errorf("database error: %s", msg)
		}
		return fmt.Errorf("failed to save initial progress")
	}

	return nil
}

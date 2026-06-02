package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SyncUser synchronizes the Google user with public.users and user_progress
func (h *Handler) SyncUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "required token"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	userData, err := h.fetchSupabaseUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	userID, email, firstName, lastName, err := extractUserData(userData)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//	fmt.Printf("DEBUG: ID=%s, Email=%s, Name=%s %s\n", userID, email, firstName, lastName)

	exists, err := h.userExists(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while verifying the user"})
		return
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{
			"synced":  false,
			"message": "usuario ya existe",
		})
		return
	}

	if err := h.createUserProfileSync(userID, email, firstName, lastName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while saving the profile"})

		return
	}

	if err := h.createUserProgress(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"synced":     true,
		"id":         userID,
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
	})
}

//
// 🔹 HELPERS
//

func (h *Handler) fetchSupabaseUser(token string) (map[string]any, error) {
	req, _ := http.NewRequest(http.MethodGet, h.SupabaseURL+"/auth/v1/user", nil)
	req.Header.Set("apikey", h.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connecting to auth service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid or expired token")
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding user data")
	}

	return data, nil
}

func extractUserData(userData map[string]any) (userID, email, firstName, lastName string, err error) {
	var ok bool
	userID, ok = userData["id"].(string)
	if !ok || userID == "" {
		return "", "", "", "", fmt.Errorf("user id not found in token")
	}

	email, _ = userData["email"].(string)

	if meta, ok := userData["user_metadata"].(map[string]any); ok {
		fullName, _ := meta["full_name"].(string)
		if fullName == "" {
			fullName, _ = meta["name"].(string)
		}

		if fullName != "" {
			parts := strings.SplitN(fullName, " ", 2)
			firstName = parts[0]
			if len(parts) > 1 {
				lastName = parts[1]
			}
		}
	}

	return userID, email, firstName, lastName, nil
}

func (h *Handler) userExists(userID string) (bool, error) {
	req, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/rest/v1/users?id=eq.%s&select=id", h.SupabaseURL, userID),
		nil,
	)

	req.Header.Set("apikey", h.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+h.SupabaseKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error connecting to database")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error verifying user existence (status %d)", resp.StatusCode)
	}

	var existing []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&existing); err != nil {
		return false, fmt.Errorf("error decoding database response")
	}

	return len(existing) > 0, nil
}

func (h *Handler) createUserProfileSync(userID, email, firstName, lastName string) error {
	username := strings.Split(email, "@")[0]
	body, _ := json.Marshal(map[string]string{
		"id":         userID,
		"first_name": firstName,
		"last_name":  lastName,
		"username":   username,
		"email":      email,
		"role":       "user",
	})

	req, _ := http.NewRequest(http.MethodPost, h.SupabaseURL+"/rest/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", h.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+h.SupabaseKey)
	req.Header.Set("Prefer", "return=representation")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error while creating profile")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return fmt.Errorf("failed to create profile (status %d)", resp.StatusCode)
		}

		// If it's a duplicate error (Postgres code 23505)
		if code, ok := errBody["code"].(string); ok && code == "23505" {
			return nil
		}

		if msg, ok := errBody["message"].(string); ok && msg != "" {
			return fmt.Errorf("database error: %s", msg)
		}
		return fmt.Errorf("failed to create profile (status %d)", resp.StatusCode)
	}
	return nil
}

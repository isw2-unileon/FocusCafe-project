package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SyncUser sincroniza el usuario de Google con public.users y user_progress
func (h *Handler) SyncUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	userData, err := h.fetchSupabaseUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
		return
	}

	userID, email, firstName, lastName, err := extractUserData(userData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	exists, err := h.userExists(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al comprobar el usuario"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar el perfil"})
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
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid token")
	}
	defer resp.Body.Close()

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func extractUserData(userData map[string]any) (userID, email, firstName, lastName string, err error) {
	userID, _ = userData["id"].(string)
	if userID == "" {
		return "", "", "", "", fmt.Errorf("missing user id")
	}

	email, _ = userData["email"].(string)

	if meta, ok := userData["user_metadata"].(map[string]any); ok {
		firstName, _ = meta["given_name"].(string)
		lastName, _ = meta["family_name"].(string)
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
		return false, err
	}
	defer resp.Body.Close()

	var existing []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&existing); err != nil {
		return false, err
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
		"provider":   "google",
	})

	req, _ := http.NewRequest(http.MethodPost, h.SupabaseURL+"/rest/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", h.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+h.SupabaseKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error creating profile")
	}
	defer resp.Body.Close()

	return nil
}

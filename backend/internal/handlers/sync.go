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
	email, _ = userData["email"].(string)

	if userID == "" {
		return "", "", "", "", fmt.Errorf("missing user id")
	}

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
	})

	req, _ := http.NewRequest(http.MethodPost, h.SupabaseURL+"/rest/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", h.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+h.SupabaseKey)
	req.Header.Set("Prefer", "return=representation")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de red: %w", err)
	}
	defer resp.Body.Close()

	// if resp.StatusCode != http.StatusCreated {
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errBody map[string]any

		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return fmt.Errorf("error decoding supabase response: %w", err)
		}
		fmt.Printf("Error Supabase: %v\n", errBody)
		return fmt.Errorf("supabase %d: %v", resp.StatusCode, errBody) // ← ahora ves el motivo
	}
	return nil
	/*
		// 1. Si el status es 201 (Creado), todo salió perfecto a la primera.
		if resp.StatusCode == http.StatusCreated {
			return nil
		}

		// 2. Si NO es 201, leemos el cuerpo para ver si es un duplicado
		var errBody map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return fmt.Errorf("error al decodificar error de supabase: %w", err)
		}

		// 3. AQUÍ AÑADES EL IF MÁGICO: <--- NUEVO
		// Buscamos el código 23505 (que es "Unique Violation" en Postgres)
		if code, ok := errBody["code"].(string); ok && code == "23505" {
			fmt.Printf("Usuario duplicado detectado para %s. Ignorando error...\n", email)
			return nil // Devolvemos nil (éxito) porque el usuario ya está en la BD
		}

		// 4. Si llegamos aquí, es un error real (ej: 400, 401, 500)
		return fmt.Errorf("supabase %d: %v", resp.StatusCode, errBody)*/
}

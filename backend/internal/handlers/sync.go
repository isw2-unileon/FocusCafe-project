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
	// 1. Obtener el token del header Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// 2. Obtener los datos del usuario desde Supabase Auth
	userReq, _ := http.NewRequest(http.MethodGet, h.SupabaseURL+"/auth/v1/user", nil)
	userReq.Header.Set("apikey", h.SupabaseKey)
	userReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil || userResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
		return
	}
	defer userResp.Body.Close()

	var userData map[string]any
	json.NewDecoder(userResp.Body).Decode(&userData)

	userID, ok := userData["id"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo obtener el ID del usuario"})
		return
	}

	email, _ := userData["email"].(string)

	// Extraer nombre y apellido del perfil de Google
	firstName := ""
	lastName := ""
	if meta, ok := userData["user_metadata"].(map[string]any); ok {
		firstName, _ = meta["given_name"].(string)
		lastName, _ = meta["family_name"].(string)
	}
	username := strings.Split(email, "@")[0]

	// 3. Comprobar si el usuario ya existe en public.users
	checkReq, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/rest/v1/users?id=eq.%s&select=id", h.SupabaseURL, userID),
		nil,
	)
	checkReq.Header.Set("apikey", h.SupabaseKey)
	checkReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.SupabaseKey))

	checkResp, err := http.DefaultClient.Do(checkReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al comprobar el usuario"})
		return
	}
	defer checkResp.Body.Close()

	var existing []map[string]any
	json.NewDecoder(checkResp.Body).Decode(&existing)

	// Si ya existe no hacemos nada
	if len(existing) > 0 {
		c.JSON(http.StatusOK, gin.H{"synced": false, "message": "usuario ya existe"})
		return
	}

	// 4. Insertar en public.users
	profileBody, _ := json.Marshal(map[string]string{
		"id":         userID,
		"first_name": firstName,
		"last_name":  lastName,
		"username":   username,
		"email":      email,
		"role":       "user",
		"provider":   "google",
	})

	profileReq, _ := http.NewRequest(
		http.MethodPost,
		h.SupabaseURL+"/rest/v1/users",
		bytes.NewBuffer(profileBody),
	)
	profileReq.Header.Set("Content-Type", "application/json")
	profileReq.Header.Set("apikey", h.SupabaseKey)
	profileReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.SupabaseKey))
	profileReq.Header.Set("Prefer", "return=representation")

	profileResp, err := http.DefaultClient.Do(profileReq)
	if err != nil || profileResp.StatusCode != http.StatusCreated {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar el perfil"})
		return
	}
	defer profileResp.Body.Close()

	// 5. Insertar en public.user_progress
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

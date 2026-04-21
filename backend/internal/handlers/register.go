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

// Register es el handler principal, solo orquesta
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cuerpo de solicitud inválido"})
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

	if err := h.createUserProfile(userID, req); err != nil {
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

// validateRegisterRequest valida los campos del formulario
func validateRegisterRequest(req *RegisterRequest) error {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)

	if req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("nombre y apellido son obligatorios")
	}
	if req.Email == "" {
		return fmt.Errorf("email es obligatorio")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("la contraseña debe tener al menos 6 caracteres")
	}
	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("las contraseñas no coinciden")
	}
	return nil
}

// createAuthUser crea el usuario en Supabase Auth y devuelve su UUID
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
		return "", fmt.Errorf("error al conectar con Supabase Auth")
	}
	defer resp.Body.Close()

	var data map[string]any
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		msg := "error al crear el usuario"
		if errMsg, ok := data["msg"].(string); ok {
			msg = errMsg
		}
		return "", fmt.Errorf("%s", msg)
	}

	return extractUserID(data)
}

// extractUserID extrae el UUID del usuario de la respuesta de Supabase
func extractUserID(data map[string]any) (string, error) {
	userMap, ok := data["user"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("respuesta inesperada de Supabase Auth")
	}

	userID, ok := userMap["id"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("no se pudo obtener el ID del usuario")
	}

	return userID, nil
}

// createUserProfile inserta el perfil del usuario en la tabla public.users
func (h *Handler) createUserProfile(userID string, req RegisterRequest) error {
	username := strings.Split(req.Email, "@")[0]

	body, _ := json.Marshal(map[string]string{
		"id":         userID,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"username":   username,
		"email":      req.Email,
		"role":       "user",
	})

	profileReq, _ := http.NewRequest(
		http.MethodPost,
		h.SupabaseURL+"/rest/v1/users",
		bytes.NewBuffer(body),
	)
	profileReq.Header.Set("Content-Type", "application/json")
	profileReq.Header.Set("apikey", h.SupabaseKey)
	profileReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.SupabaseKey))
	profileReq.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(profileReq)
	if err != nil {
		return fmt.Errorf("usuario creado en auth pero falló el perfil")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var profileErr map[string]any
		err := json.NewDecoder(resp.Body).Decode(&profileErr)
		if err != nil {
			return err
		}
		return fmt.Errorf("error al guardar el perfil")
	}

	return nil
}

// createUserProgress inserta el progreso inicial del usuario en public.user_progress
func (h *Handler) createUserProgress(userID string) error {
	body, _ := json.Marshal(map[string]any{
		"user_id": userID,
		"energy":  500,
		"level":   1,
	})

	progressReq, _ := http.NewRequest(
		http.MethodPost,
		h.SupabaseURL+"/rest/v1/user_progress",
		bytes.NewBuffer(body),
	)
	progressReq.Header.Set("Content-Type", "application/json")
	progressReq.Header.Set("apikey", h.SupabaseKey)
	progressReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.SupabaseKey))
	progressReq.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(progressReq)
	if err != nil {
		return fmt.Errorf("error al crear el progreso del usuario")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var progressErr map[string]any
		err := json.NewDecoder(resp.Body).Decode(&progressErr)
		if err != nil {
			return err
		}
		return fmt.Errorf("error al guardar el progreso")
	}

	return nil
}

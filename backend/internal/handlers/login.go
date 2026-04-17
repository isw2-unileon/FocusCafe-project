package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	SupabaseURL string
	SupabaseKey string
	Auth        auth.TokenValidator
}

// Login es el handler principal, solo orquesta
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	token, user, err := h.authenticateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// authenticateUser orquesta el proceso de autenticación
func (h *Handler) authenticateUser(email, password string) (string, interface{}, error) {
	body, err := buildLoginBody(email, password)
	if err != nil {
		return "", nil, fmt.Errorf("error al construir la petición")
	}

	resp, err := h.callSupabaseAuth(body)
	if err != nil {
		return "", nil, fmt.Errorf("error al conectar con Supabase")
	}
	defer resp.Body.Close()

	return parseAuthResponse(resp)
}

// buildLoginBody construye el cuerpo JSON de la petición
func buildLoginBody(email, password string) ([]byte, error) {
	return json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
}

// callSupabaseAuth hace la llamada HTTP a Supabase
func (h *Handler) callSupabaseAuth(body []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("POST",
		h.SupabaseURL+"/auth/v1/token?grant_type=password",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", h.SupabaseKey)

	client := &http.Client{}
	return client.Do(httpReq)
}

// GoogleAuth redirige al proveedor de Google a través de Supabase
func (h *Handler) GoogleAuth(c *gin.Context) {
	redirectURL := fmt.Sprintf(
		"%s/auth/v1/authorize?provider=google&redirect_to=http://localhost:5173/home",
		h.SupabaseURL,
	)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// parseAuthResponse procesa la respuesta de Supabase y extrae el token y el usuario
func parseAuthResponse(resp *http.Response) (string, interface{}, error) {
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != http.StatusOK {
		errMsg, ok := result["error_description"].(string)
		if !ok || errMsg == "" {
			errMsg = "Credenciales incorrectas"
		}
		return "", nil, fmt.Errorf("%s", errMsg)
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", nil, fmt.Errorf("error al obtener el token")
	}

	return token, result["user"], nil
}

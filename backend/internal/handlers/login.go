package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoginRequest defines the required info for the authentication
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login is the principal handler, it only orquests
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

// authenticateUser orquests authentication process
func (h *Handler) authenticateUser(email, password string) (string, interface{}, error) {
	body, err := buildLoginBody(email, password)
	if err != nil {
		return "", nil, fmt.Errorf("error creating the request")
	}

	resp, err := h.callSupabaseAuth(body)
	if err != nil {
		return "", nil, fmt.Errorf("error connecting to Supabase")
	}
	defer resp.Body.Close()

	return parseAuthResponse(resp)
}

// buildLoginBody constructs petition's JSON body
func buildLoginBody(email, password string) ([]byte, error) {
	return json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
}

// callSupabaseAuth makes the HTTP call to Supabase
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

// GoogleAuth redirects the user to the Google provider via Supabase.
func (h *Handler) GoogleAuth(c *gin.Context) {
	redirectURL := fmt.Sprintf(
		"%s/auth/v1/authorize?provider=google&redirect_to=http://localhost:5173/home",
		h.SupabaseURL,
	)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// parseAuthResponse process the supabase's response and extracts token and user
func parseAuthResponse(resp *http.Response) (string, interface{}, error) {
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, fmt.Errorf("error at the codifying the response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg, ok := result["error_description"].(string)
		if !ok || errMsg == "" {
			errMsg = "Incorrect credentials"
		}
		return "", nil, fmt.Errorf("%s", errMsg)
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", nil, fmt.Errorf("error retrieving the token")
	}

	return token, result["user"], nil
}

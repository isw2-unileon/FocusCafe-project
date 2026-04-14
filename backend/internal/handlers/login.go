package handlers

import "github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"

// Handler encapsulates the logic for user login requests.
type Handler struct {
	SupabaseURL string
	SupabaseKey string
	Auth        auth.TokenValidator
}

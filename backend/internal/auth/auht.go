package auth

// UserClaims contains info extracted from the token
type UserClaims struct {
	Email string
	ID    string
}

// TokenValidator defines the interface for validating tokens
type TokenValidator interface {
	ValidateToken(tokenStr string) (*UserClaims, error)
}

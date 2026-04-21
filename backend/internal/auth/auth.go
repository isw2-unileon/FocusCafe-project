package auth

import "github.com/golang-jwt/jwt/v5"

// UserClaims contains info extracted from the token
type UserClaims struct {
	jwt.RegisteredClaims

	Email string `json:"email"`
	Role  string `json:"role"`
	ID    string `json:"-"`
}

// TokenValidator defines the interface for validating tokens
type TokenValidator interface {
	ValidateToken(tokenStr string) (*UserClaims, error)
}

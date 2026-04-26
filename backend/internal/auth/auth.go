package auth

import "github.com/golang-jwt/jwt/v5"

// UserClaims contains info extracted from the token.
// We embed RegisteredClaims to get standard fields like Subject (ID), ExpiresAt, etc.
type UserClaims struct {
	jwt.RegisteredClaims
	Email       string      `json:"email"`
	AppMetadata AppMetadata `json:"app_metadata"`
}

// AppMetadata contains custom fields related to the user, such as their role.
type AppMetadata struct {
	Role string `json:"role"`
}

// GetID returns the user ID from the Subject claim
func (c *UserClaims) GetID() string {
	return c.Subject
}

// IsAdmin checks if the user has the admin role
func (c *UserClaims) IsAdmin() bool {
	return c.AppMetadata.Role == "admin"
}

// TokenValidator defines the interface for validating tokens
type TokenValidator interface {
	ValidateToken(tokenStr string) (*UserClaims, error)
}

package auth

import "github.com/golang-jwt/jwt/v5"

// UserClaims contains info extracted from the token
type UserClaims struct {
	// Esto es OBLIGATORIO para que el validador automático funcione.
	// Incluye internamente: iss, sub, aud, exp, nbf, iat, jti
	jwt.RegisteredClaims

	// Estos son tus campos personalizados de Supabase
	Email string `json:"email"`
	Role  string `json:"role"`

	// El ID de Supabase viene en el campo "sub" (Subject) de RegisteredClaims.
	// Si necesitas usar el nombre "ID", lo mapearemos en el validador.
	ID string `json:"-"`
}

// TokenValidator defines the interface for validating tokens
type TokenValidator interface {
	ValidateToken(tokenStr string) (*UserClaims, error)
}

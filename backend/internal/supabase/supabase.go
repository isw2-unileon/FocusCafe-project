package supabase

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
)

// JWTAdapter handles Supabase JWT verification
type JWTAdapter struct {
	publicKey *ecdsa.PublicKey
}

// NewJWTAdapter creates a new instance of JWTAdpater
func NewJWTAdapter(pemString string) (*JWTAdapter, error) {
	// Parse public key
	key, err := jwt.ParseECPublicKeyFromPEM([]byte(pemString))
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}
	return &JWTAdapter{publicKey: key}, nil
}

// ValidateToken checks the validity of a given JWT string
func (a *JWTAdapter) ValidateToken(tokenString string) (*auth.UserClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verification of method ES256
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signature method %v", token.Header["alg"])
		}
		return a.publicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("error in claims")
	}

	email, ok := claims["email"].(string)
	if !ok {
		email = ""
	}

	id, ok := claims["sub"].(string)

	if !ok {
		return nil, fmt.Errorf("token doesn't contain userId")
	}
	return &auth.UserClaims{
		Email: email,
		ID:    id,
	}, nil
}

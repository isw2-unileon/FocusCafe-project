package supabase

import (
	"context"
	"fmt"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
)

// JWTAdapter handles Supabase JWT verification
type JWTAdapter struct {
	jwks keyfunc.Keyfunc
}

// NewJWTAdapter creates a new instance of JWTAdapter
func NewJWTAdapter(supabaseURL string) (*JWTAdapter, error) {
	if supabaseURL == "" {
		return nil, fmt.Errorf("supabase URL is required")
	}

	jwksURL := supabaseURL + "/auth/v1/.well-known/jwks.json"

	// Search for the key
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Download the key
	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keys from Supabase: %w", err)
	}
	return &JWTAdapter{jwks: jwks}, nil
}

// ValidateToken checks the validity of a given JWT string
func (a *JWTAdapter) ValidateToken(tokenString string) (*auth.UserClaims, error) {
	claims := &auth.UserClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		a.jwks.Keyfunc,
		jwt.WithValidMethods([]string{"HS256", "ES256", "RS256"}),
	)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token invalid signature")
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("token doesn't contain userId (sub)")
	}

	return claims, nil
}

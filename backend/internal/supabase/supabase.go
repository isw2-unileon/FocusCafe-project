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

// NewJWTAdapter creates a new instance of JWTAdpater
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
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"ES256", "RS256"}),
	)

	token, err := parser.Parse(tokenString, a.jwks.Keyfunc)
	if err != nil {
		fmt.Printf("DEBUG: Error at parsing: %v\n", err)
		return nil, err
	}

	if !token.Valid {
		fmt.Println("False signature")
		return nil, fmt.Errorf("token invalid signature")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("error in claims")
	}
	fmt.Printf("Extracted claims: %v\n", claims)

	id, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	// // Extract rol frfom appMetadata
	// var role string
	// if appMetadata, ok := claims["app_metadata"].(map[string]any); ok {
	// 	role, _ = appMetadata["role"].(string)
	// }

	if id == "" {
		return nil, fmt.Errorf("token doesn't contain userId (sub)")
	}

	return &auth.UserClaims{
		ID:    id,
		Email: email,
		Role:  role,
	}, nil
}

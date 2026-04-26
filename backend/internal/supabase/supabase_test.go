package supabase_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/supabase"
)

func TestValidateToken(t *testing.T) {
	// 1. Generate private key
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// 2. Create a test server (JWKS Mock)
	// It simulates being Supabase delivering the publick key
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x := base64.RawURLEncoding.EncodeToString(privateKey.X.Bytes())
		y := base64.RawURLEncoding.EncodeToString(privateKey.Y.Bytes())

		fmt.Fprintf(w, `{
            "keys": [
                {
                    "kty": "EC",
                    "crv": "P-256",
                    "x": "%s",
                    "y": "%s",
                    "alg": "ES256",
                    "kid": "test-key"
                }
            ]
        }`, x, y)
	}))
	defer server.Close()

	// 3. Generate a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub":   "1234567890",
		"email": "test@focuscafe.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iss":   server.URL + "/auth/v1",
		"aud":   "authenticated",
	})
	// Add kid to header
	token.Header["kid"] = "test-key"
	validToken, _ := token.SignedString(privateKey)

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
	}{
		{
			name:        "Valid Token",
			tokenString: validToken,
			wantErr:     false,
		},
		{
			name:        "Malformed Token",
			tokenString: "not.a.token",
			wantErr:     true,
		},
		{
			name:        "Invalid Signature",
			tokenString: validToken + "corrupt",
			wantErr:     true,
		},
	}

	adapter, err := supabase.NewJWTAdapter(server.URL)
	if err != nil {
		t.Fatalf("No se pudo crear el adaptador: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := adapter.ValidateToken(tt.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

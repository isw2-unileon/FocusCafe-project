package supabase_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/supabase"
)

func TestValidateToken(t *testing.T) {
	// Generate a private key for the tests
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// Generate a public key
	pubBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
	t.Log(pubPEM)

	// Generate valid token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub":   "1234567890",
		"email": "test@focuscafe.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	validToken, _ := token.SignedString(privateKey)
	t.Log(validToken)

	tests := []struct {
		name          string
		tokenString   string
		publicKey     string
		wantErr       bool
		expectedError string
	}{
		{
			name:        "Token with wrong signature",
			tokenString: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.e30.signature",
			publicKey:   pubPEM,
			wantErr:     true,
		},
		{
			name:        "Invalid signature",
			tokenString: "eyJhbGciOiJFUzI1NiJ9.e30.badsignature",
			publicKey:   pubPEM,
			wantErr:     true,
		},
		{
			name:        "Malformed Token",
			tokenString: "not.a.token",
			publicKey:   pubPEM,
			wantErr:     true,
		},
		{
			name:          "Wrong algorithm",
			tokenString:   "eyJhbGciOiJIUzI1NiJ9.e30.signature", // HS256
			publicKey:     pubPEM,
			wantErr:       true,
			expectedError: "invalid token",
		},
		{
			name:        "Valid Token",
			tokenString: validToken,
			publicKey:   pubPEM,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := supabase.NewJWTAdapter(tt.publicKey)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Critical error: %v", err)
				}
				return
			}

			_, err = adapter.ValidateToken(tt.tokenString)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.expectedError != "" && err != nil {
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error: %v, but got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

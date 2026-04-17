package supabase_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	//"crypto/x509"
	"encoding/base64"
	//"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/supabase"
)

// func TestValidateToken(t *testing.T) {
// 	// Generate a private key for the tests
// 	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

// 	// Generate a public key
// 	pubBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
// 	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))

// 	// Generate valid token
// 	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
// 		"sub":   "1234567890",
// 		"email": "test@focuscafe.com",
// 		"exp":   time.Now().Add(time.Hour).Unix(),
// 	})
// 	validToken, _ := token.SignedString(privateKey)

// 	tests := []struct {
// 		name          string
// 		tokenString   string
// 		publicKey     string
// 		wantErr       bool
// 		expectedError string
// 	}{
// 		{
// 			name:        "Token with wrong signature",
// 			tokenString: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.e30.signature",
// 			publicKey:   pubPEM,
// 			wantErr:     true,
// 		},
// 		{
// 			name:        "Invalid signature",
// 			tokenString: "eyJhbGciOiJFUzI1NiJ9.e30.badsignature",
// 			publicKey:   pubPEM,
// 			wantErr:     true,
// 		},
// 		{
// 			name:        "Malformed Token",
// 			tokenString: "not.a.token",
// 			publicKey:   pubPEM,
// 			wantErr:     true,
// 		},
// 		{
// 			name:          "Wrong algorithm",
// 			tokenString:   "eyJhbGciOiJIUzI1NiJ9.e30.signature", // HS256
// 			publicKey:     pubPEM,
// 			wantErr:       true,
// 			expectedError: "invalid token",
// 		},
// 		{
// 			name:        "Valid Token",
// 			tokenString: validToken,
// 			publicKey:   pubPEM,
// 			wantErr:     false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			adapter, err := supabase.NewJWTAdapter(tt.publicKey)
// 			if err != nil {
// 				if !tt.wantErr {
// 					t.Fatalf("Critical error: %v", err)
// 				}
// 				return
// 			}

// 			_, err = adapter.ValidateToken(tt.tokenString)

// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			if tt.expectedError != "" && err != nil {
// 				if !strings.Contains(err.Error(), tt.expectedError) {
// 					t.Errorf("Expected error: %v, but got: %v", tt.expectedError, err)
// 				}
// 			}
// 		})
// 	}
// }

func TestValidateToken(t *testing.T) {
	// 1. Generamos la clave privada para firmar los tokens del test
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// 2. CREAMOS UN SERVIDOR DE PRUEBAS (JWKS Mock)
	// Esto simula ser Supabase entregando la llave pública en formato JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extraemos las coordenadas X e Y de la clave pública para el formato JWK
		x := base64.RawURLEncoding.EncodeToString(privateKey.X.Bytes())
		y := base64.RawURLEncoding.EncodeToString(privateKey.Y.Bytes())

		// Este es el JSON que tu NewJWTAdapter irá a buscar
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

	// 3. Generamos un token VÁLIDO firmado con nuestra clave privada
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub":   "1234567890",
		"email": "test@focuscafe.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iss":   server.URL + "/auth/v1", // Importante: el emisor debe coincidir
		"aud":   "authenticated",
	})
	// Añadimos el KID al header para que el validador encuentre la llave en el mock
	token.Header["kid"] = "test-key"
	validToken, _ := token.SignedString(privateKey)

	// 4. ESTRUCTURA DE TESTS
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

	// 5. INICIALIZAMOS EL ADAPTADOR apuntando a nuestro servidor local
	adapter, err := supabase.NewJWTAdapter(server.URL) // Le pasamos la URL del mock
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

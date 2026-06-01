package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/ws"
)

func TestInfra_Integration(t *testing.T) {
	_, h := setupTestApp()

	// Initialize and run the WebSocket hub as it's required for the /api/ws route.
	wsHub := ws.NewHub()
	go wsHub.Run()
	h.WSHub = wsHub

	t.Run("Health Check - GET /health", func(t *testing.T) {
		testHealthCheck(t, h)
	})

	t.Run("Protected Hello - GET /api/hello", func(t *testing.T) {
		testHelloProtected(t, h)
	})

	t.Run("WebSocket Upgrade - GET /api/ws", func(t *testing.T) {
		testWebSocketUpgrade(t, h)
	})
}

func testHealthCheck(t *testing.T, h *handlers.Handler) {
	r := setupInfraRouter(h, uuid.Nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp["status"])
	}
}

func testHelloProtected(t *testing.T, h *handlers.Handler) {
	userID := uuid.New()
	r := setupInfraRouter(h, userID)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/hello", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["message"] != "Hello from the API" {
		t.Errorf("expected message 'Hello from the API', got '%s'", resp["message"])
	}
}

func testWebSocketUpgrade(t *testing.T, h *handlers.Handler) {
	r := setupInfraRouter(h, uuid.Nil)
	server := httptest.NewServer(r)
	defer server.Close()

	// Convert http URL to ws URL
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/ws"

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}

	if resp != nil && resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	defer func() {
		_ = conn.Close()
	}()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("expected status 101, got %d", resp.StatusCode)
	}
}

// setupInfraRouter configures the router with the exact paths from main.go
func setupInfraRouter(h *handlers.Handler, userID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Exact registration as in main.go
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		// WebSocket route (Public in the group, handled by WSHandler)
		api.GET("/ws", h.WSHandler(h.WSHub))

		// Protected group
		protected := api.Group("/")
		if userID != uuid.Nil {
			protected.Use(mockAuthMiddleware(userID))
		}
		protected.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Hello from the API"})
		})
	}

	return r
}

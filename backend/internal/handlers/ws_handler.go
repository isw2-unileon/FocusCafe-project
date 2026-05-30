package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, we should check against cfg.CORSAllowOrigin
	},
}

// WSHandler handles websocket requests from the peer.
func (h *Handler) WSHandler(hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Upgrade connection without immediate auth (standard WS doesn't support headers)
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		// 2. Create client but don't register yet
		client := &ws.Client{
			Hub:  hub,
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		// 3. Start pumps
		go client.WritePump()
		// Pass dependencies to ReadPump for validation
		go client.ReadPump(h.Auth, h.UserService)
	}
}

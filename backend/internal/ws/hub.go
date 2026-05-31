package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	service "github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

// Message defines the structure of data sent over WebSocket
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client represents a single WebSocket connection
type Client struct {
	Hub     *Hub
	Conn    *websocket.Conn
	Send    chan []byte
	UserID  uuid.UUID
	GroupID *int64
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	mu sync.RWMutex
}

// NewHub creates and initializaes a new instance of Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run executes the main loop of the Hub to manage client lifecycles and broadcasting
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("Client unregistered: %s", client.UserID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterClient registers a new client with the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// BroadcastToGroup sends a message to all users in a specific group
func (h *Hub) BroadcastToGroup(groupID int64, msg Message) {
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.GroupID != nil && *client.GroupID == groupID {
			select {
			case client.Send <- data:
			default:
				// Prevents blocking if client channel is full
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uuid.UUID, msg Message) {
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				// Prevents blocking if client channel is full
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	defer func() {
		_ = c.Conn.Close()
	}()
	for message := range c.Send {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		_, _ = w.Write(message) // Explicitly ignore error to satisfy errcheck

		if err := w.Close(); err != nil {
			return
		}
	}
	// If the channel is closed by the hub
	_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// ReadPump pumps messages from the WebSocket coneection to the hub
func (c *Client) ReadPump(validator auth.TokenValidator, userService service.UserServiceInterface) {
	defer func() {
		c.Hub.unregister <- c
		_ = c.Conn.Close()
	}()

	authenticated := false

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if !authenticated {
			if c.handleAuthentication(message, validator, userService) {
				authenticated = true
			}
			continue
		}

		// Handle other messages from client if needed
	}
}

// handleAuthentication is a privagte helper that proceses the AUTH handshake message
func (c *Client) handleAuthentication(message []byte, validator auth.TokenValidator, userService service.UserServiceInterface) bool {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return false
	}
	if msg.Type != "AUTH" {
		return false
	}

	token, ok := msg.Payload.(string)
	if !ok {
		return false
	}

	claims, err := validator.ValidateToken(token)
	if err != nil {
		log.Printf("Auth failed: %v", err)
		return false
	}

	userID, _ := uuid.Parse(claims.GetID())
	c.UserID = userID

	profile, err := userService.GetUserProfile(context.Background(), userID)
	if err == nil && profile.Group != nil {
		groupID := profile.Group.ID
		c.GroupID = &groupID
	}

	c.Hub.RegisterClient(c)
	log.Printf("Client authenticated and registered: %s", c.UserID)
	return true
}

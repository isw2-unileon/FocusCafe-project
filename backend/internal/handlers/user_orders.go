package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/ws"
)

// GetUserOrders obtains the orders of the authenticated user.
func (h *Handler) GetUserOrders(c *gin.Context) {
	// Obtain user ID from JWT claims
	id, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Obtain user orders from the service layer
	orders, err := h.UserOrdersService.GetUserOrders(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user orders"})
		return
	}

	// Return user orders as JSON response
	c.JSON(http.StatusOK, orders)
}

// CompleteUserOrder handles completing an order
func (h *Handler) CompleteUserOrder(c *gin.Context) {
	// 1. Extract id from url
	idParam := c.Param("id")
	orderID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = h.UserOrdersService.CompleteUserOrder(c.Request.Context(), userID, uint(orderID))
	if err != nil {

		if err.Error() == "insufficient energy" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough energy"})
			return
		}
		if err.Error() == "order already completed" {
			c.JSON(http.StatusConflict, gin.H{"error": "This order has already been completed by another user"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error at completing the order: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order successfully completed!",
		"status":  "completed",
	})

	if h.WSHub != nil {
		// 4. Notify via WebSocket
		go func() {
			// Use Background context because the request context will be cancelled
			bgCtx := context.Background()

			h.WSHub.SendToUser(userID, ws.Message{
				Type:    "ORDERS_UPDATED",
				Payload: gin.H{"order_id": orderID},
			})

			// If the user belongs to a group, also notify the group
			profile, err := h.UserService.GetUserProfile(bgCtx, userID)
			if err != nil {
				log.Printf("Error fetching user profile for WS broadcast: %v", err)
				return
			}

			if profile.Group != nil {
				h.WSHub.BroadcastToGroup(profile.Group.ID, ws.Message{
					Type:    "ORDERS_UPDATED",
					Payload: gin.H{"order_id": orderID, "completed_by": userID},
				})
			}
		}()
	} else {
		log.Println("Skipping WebSocket broadcast: WSHub is nil")
	}
}

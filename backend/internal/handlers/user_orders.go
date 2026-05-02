package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

func (h *Handler) CompleteUserOrder(c *gin.Context) {
	fmt.Printf("buenas")
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

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error at completing the order: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order succesfully completed!",
		"status":  "completed",
	})
}

package handlers

import (
	"IAM-server/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UpdateNotifyTokenRequest struct {
	NotifyToken string `json:"notify_token" binding:"required"`
}

// UpdateNotifyTokenHandler updates the push notification token for the authenticated customer.
// @Summary Update customer notify token
// @Description Updates the FCM/APNs push notification token associated with the authenticated customer's account.
// @Tags Customer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateNotifyTokenRequest true "Notify token payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customer/notify-token [put]
func UpdateNotifyTokenHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateNotifyTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := strings.TrimSpace(req.NotifyToken)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notify_token cannot be empty"})
		return
	}

	if err := services.UpdateNotifyToken(c.Request.Context(), user.ID, &token); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notify token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notify token updated successfully",
	})
}

// DeleteNotifyTokenHandler clears the push notification token for the authenticated customer.
// @Summary Remove customer notify token
// @Description Clears the stored push notification token when customer logs out or unregisters from push notifications.
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customer/notify-token [delete]
func DeleteNotifyTokenHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := services.UpdateNotifyToken(c.Request.Context(), user.ID, nil); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear notify token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notify token cleared successfully",
	})
}

// GetCustomerBroadcastMessagesHandler retrieves broadcast memos from subscribed vendor apps.
// @Summary Get customer broadcast messages
// @Description Returns broadcast messages sent by apps that the authenticated customer is subscribed to.
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Param app_id query string false "Filter by specific App ID"
// @Param limit query int false "Number of messages to return (default 50, max 100)"
// @Param offset query int false "Pagination offset (default 0)"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customer/messages/broadcast [get]
func GetCustomerBroadcastMessagesHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	appID := strings.TrimSpace(c.Query("app_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	memos, err := services.GetCustomerBroadcastMessages(c.Request.Context(), user.ID, appID, limit, offset)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve broadcast messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": memos,
		"count":    len(memos),
		"limit":    limit,
		"offset":   offset,
	})
}

// GetCustomerDirectMessagesHandler retrieves 1-to-1 direct memos sent to the customer.
// @Summary Get customer direct messages
// @Description Returns direct unicast messages sent specifically to the authenticated customer by vendors.
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Param app_id query string false "Filter by specific App ID"
// @Param limit query int false "Number of messages to return (default 50, max 100)"
// @Param offset query int false "Pagination offset (default 0)"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customer/messages/direct [get]
func GetCustomerDirectMessagesHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	appID := strings.TrimSpace(c.Query("app_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	memos, err := services.GetCustomerDirectMessages(c.Request.Context(), user.ID, appID, limit, offset)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve direct messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": memos,
		"count":    len(memos),
		"limit":    limit,
		"offset":   offset,
	})
}

// GetCustomerSubscriptionsHandler retrieves subscribed vendor apps for broadcast, unicast, or both.
// @Summary Get subscribed vendors
// @Description Returns all vendor apps the customer is subscribed to for broadcast and unicast notifications.
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Param type query string false "Filter subscription type: broadcast, unicast, or all (default all)"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customer/subscriptions [get]
func GetCustomerSubscriptionsHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subType := strings.ToLower(strings.TrimSpace(c.Query("type")))

	switch subType {
	case "broadcast":
		subs, err := services.GetCustomerBroadcastSubscriptions(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve broadcast subscriptions"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"broadcast": subs,
			"count":     len(subs),
		})
	case "unicast":
		subs, err := services.GetCustomerUnicastSubscriptions(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve unicast subscriptions"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"unicast": subs,
			"count":   len(subs),
		})
	default:
		subs, err := services.GetCustomerSubscriptions(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve subscriptions"})
			return
		}
		c.JSON(http.StatusOK, subs)
	}
}

// GetCustomerBroadcastSubscriptionsHandler retrieves all broadcast subscribed vendor apps for the customer.
// @Summary Get broadcast subscribed vendors
// @Description Returns all apps the customer is subscribed to for broadcast notifications.
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customer/subscriptions/broadcast [get]
func GetCustomerBroadcastSubscriptionsHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subs, err := services.GetCustomerBroadcastSubscriptions(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve broadcast subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subs,
		"count":         len(subs),
	})
}

// GetCustomerUnicastSubscriptionsHandler retrieves all unicast subscribed vendor apps for the customer.
// @Summary Get unicast subscribed vendors
// @Description Returns all apps the customer is subscribed to for direct/unicast notifications.
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customer/subscriptions/unicast [get]
func GetCustomerUnicastSubscriptionsHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subs, err := services.GetCustomerUnicastSubscriptions(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve unicast subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subs,
		"count":         len(subs),
	})
}

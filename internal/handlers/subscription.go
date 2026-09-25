package handlers

import (
	"IAM-server/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SubscribeBroadcastRequest struct {
	AppID string `json:"app_id" binding:"required"`
}

type SubscribeUnicastRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	VendorUserID string `json:"vendor_user_id" binding:"required"`
}

// SubscribeBroadcastHandler subscribes a user to an app's broadcast notifications.
// @Summary Subscribe to app broadcast
// @Tags Subscriptions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SubscribeBroadcastRequest true "App ID"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/broadcast [post]
func SubscribeBroadcastHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SubscribeBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := services.SubscribeBroadcast(c.Request.Context(), user.ID, req.AppID)
	if err != nil {
		if err.Error() == "app not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Subscribed to broadcast successfully",
		"subscription": sub,
	})
}

// UnsubscribeBroadcastHandler removes a user's subscription to an app's broadcast notifications.
// @Summary Unsubscribe from app broadcast
// @Tags Subscriptions
// @Security BearerAuth
// @Produce json
// @Param app_id path string true "App ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/broadcast/{app_id} [delete]
func UnsubscribeBroadcastHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	appID := c.Param("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id parameter is required"})
		return
	}

	if err := services.UnsubscribeBroadcast(c.Request.Context(), user.ID, appID); err != nil {
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Unsubscribed from broadcast successfully",
	})
}

// SubscribeUnicastHandler subscribes a user to an app's unicast notifications.
// @Summary Subscribe to app unicast
// @Tags Subscriptions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SubscribeUnicastRequest true "Subscription payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/unicast [post]
func SubscribeUnicastHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SubscribeUnicastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, unicastKey, err := services.SubscribeUnicast(c.Request.Context(), user.ID, req.AppID, req.VendorUserID)
	if err != nil {
		if err.Error() == "app not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe to unicast"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Subscribed to unicast successfully",
		"subscription": sub,
		"unicast_key":  unicastKey,
	})
}

// UnsubscribeUnicastHandler removes a user's unicast subscription for an app.
// @Summary Unsubscribe from app unicast
// @Tags Subscriptions
// @Security BearerAuth
// @Produce json
// @Param app_id path string true "App ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/unicast/{app_id} [delete]
func UnsubscribeUnicastHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	appID := c.Param("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id parameter is required"})
		return
	}

	if err := services.UnsubscribeUnicast(c.Request.Context(), user.ID, appID); err != nil {
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe from unicast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Unsubscribed from unicast successfully",
	})
}

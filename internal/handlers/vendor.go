package handlers

import (
	"IAM-server/internal/models"
	"IAM-server/internal/services"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type PushBroadcastRequest struct {
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
	Image   *string `json:"image,omitempty"`
}

type PushUnicastRequest struct {
	UnicastKey string  `json:"unicast_key,omitempty"`
	UserID     string  `json:"user_id,omitempty"`
	Title      string  `json:"title" binding:"required"`
	Content    string  `json:"content" binding:"required"`
	Image      *string `json:"image,omitempty"`
}

// resolveVendorApp retrieves the authenticated vendor App from the Gin context.
func resolveVendorApp(c *gin.Context) (*models.App, error) {
	val, exists := c.Get("vendor_app")
	if !exists {
		return nil, errors.New("vendor app not found in context")
	}

	app, ok := val.(*models.App)
	if !ok || app == nil {
		return nil, errors.New("invalid vendor app in context")
	}

	return app, nil
}

// PushBroadcastHandler handles broadcast message dispatch by a vendor app.
// @Summary Push broadcast message
// @Description Vendor sends a broadcast message to all subscribed users. Persists to database.
// @Tags Vendor
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body PushBroadcastRequest true "Broadcast memo payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vendor/broadcast [post]
func PushBroadcastHandler(c *gin.Context) {
	app, err := resolveVendorApp(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req PushBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	memo, err := services.PushBroadcast(c.Request.Context(), app.ID, req.Title, req.Content, req.Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to push broadcast memo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Broadcast memo pushed successfully",
		"memo":    memo,
	})
}

// PushUnicastHandler handles unicast direct message dispatch by a vendor app.
// @Summary Push unicast message
// @Description Vendor sends a direct 1-to-1 message to a specific subscribed user using unicast_key or user_id. Persists to database.
// @Tags Vendor
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body PushUnicastRequest true "Unicast memo payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vendor/unicast [post]
func PushUnicastHandler(c *gin.Context) {
	app, err := resolveVendorApp(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req PushUnicastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	memo, err := services.PushUnicast(c.Request.Context(), app.ID, req.UnicastKey, req.UserID, req.Title, req.Content, req.Image)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
			return
		}
		if strings.Contains(errMsg, "must be provided") || strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to push unicast memo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Unicast memo pushed successfully",
		"memo":    memo,
	})
}

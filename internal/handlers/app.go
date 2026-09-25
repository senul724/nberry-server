package handlers

import (
	"IAM-server/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CreateAppRequest struct {
	Name               string  `json:"name" binding:"required"`
	Description        string  `json:"description"`
	Logo               *string `json:"logo"`
	UnicastCallbackURL *string `json:"unicast_callback_url"`
}

// resolveAppID extracts the app ID parameter from route.
func resolveAppID(c *gin.Context) (string, bool) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app ID parameter is required"})
		return "", false
	}
	return appID, true
}

// CreateAppHandler creates a new app and generates an API secret.
// @Summary Create an application
// @Tags Apps
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateAppRequest true "App details"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps [post]
func CreateAppHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app, secretKey, err := services.CreateApp(c.Request.Context(), user.ID, req.Name, req.Description, req.Logo, req.UnicastCallbackURL)
	if err != nil {
		if strings.Contains(err.Error(), "invalid user ID") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create app"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "App created successfully",
		"app":        app,
		"secret_key": secretKey,
	})
}

// DeleteAppHandler deletes an existing app owned by the authenticated user.
// @Summary Delete an application
// @Tags Apps
// @Security BearerAuth
// @Produce json
// @Param id path string true "App ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps/{id} [delete]
func DeleteAppHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	appID, ok := resolveAppID(c)
	if !ok {
		return
	}

	if err := services.DeleteApp(c.Request.Context(), user.ID, appID); err != nil {
		if err.Error() == "app not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete app"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "App deleted successfully",
	})
}

// GetAppsHandler lists all apps owned by the authenticated user.
// @Summary List all user applications
// @Tags Apps
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps [get]
func GetAppsHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	apps, err := services.GetUserApps(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve apps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apps": apps,
	})
}

// GetAppHandler retrieves a single app by ID.
// @Summary Get application details
// @Tags Apps
// @Security BearerAuth
// @Produce json
// @Param id path string true "App ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps/{id} [get]
func GetAppHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	appID, ok := resolveAppID(c)
	if !ok {
		return
	}

	app, err := services.GetAppByID(c.Request.Context(), user.ID, appID)
	if err != nil {
		if err.Error() == "app not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve app"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"app": app,
	})
}

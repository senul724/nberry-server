package middleware

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"IAM-server/internal/services"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireVendorAuth validates the vendor's API secret key from request headers.
// It checks 'X-API-Key' first, then falls back to 'Authorization: Bearer <key>'.
func RequireVendorAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := strings.TrimSpace(c.GetHeader("X-API-Key"))

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required. Provide via 'X-API-Key' or 'Authorization: Bearer <key>' header",
			})
			c.Abort()
			return
		}

		// Calculate SHA-256 hash of the provided secret key
		secretHash := services.HashAppSecret(apiKey)

		var app models.App
		err := connections.DB.WithContext(c.Request.Context()).
			Where("secret_hash = ?", secretHash).
			First(&app).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "invalid API key",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to authenticate vendor app",
			})
			c.Abort()
			return
		}

		// Store authenticated app in Gin context
		c.Set("vendor_app", &app)
		c.Next()
	}
}

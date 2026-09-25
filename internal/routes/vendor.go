package routes

import (
	"IAM-server/internal/handlers"
	"IAM-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetVendorRoutes registers vendor endpoints on the public routes group.
// All endpoints are secured by vendor API key authentication middleware.
func SetVendorRoutes(api *gin.RouterGroup) {
	vendor := api.Group("/vendor")
	vendor.Use(middleware.RequireVendorAuth())
	{
		vendor.POST("/broadcast", handlers.PushBroadcastHandler)
		vendor.POST("/unicast", handlers.PushUnicastHandler)
	}

	// Also expose under /push for convenience
	push := api.Group("/push")
	push.Use(middleware.RequireVendorAuth())
	{
		push.POST("/broadcast", handlers.PushBroadcastHandler)
		push.POST("/unicast", handlers.PushUnicastHandler)
	}
}

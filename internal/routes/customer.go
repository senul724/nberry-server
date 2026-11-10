package routes

import (
	"IAM-server/internal/handlers"
	"IAM-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetCustomerRoutes registers routes for customer operations.
// Protected by general BearerAuth middleware.
func SetCustomerRoutes(api *gin.RouterGroup) {
	customer := api.Group("/customer")
	customer.Use(middleware.ProtectRoute())

	// Update or clear device push notification token (FCM/APNs)
	customer.PUT("/notify-token", handlers.UpdateNotifyTokenHandler)
	customer.PATCH("/notify-token", handlers.UpdateNotifyTokenHandler)
	customer.DELETE("/notify-token", handlers.DeleteNotifyTokenHandler)

	// Message feeds
	customer.GET("/messages/broadcast", handlers.GetCustomerBroadcastMessagesHandler)
	customer.GET("/messages/direct", handlers.GetCustomerDirectMessagesHandler)

	// Subscribed vendors (subscriptions)
	customer.GET("/subscriptions", handlers.GetCustomerSubscriptionsHandler)
	customer.GET("/subscriptions/broadcast", handlers.GetCustomerBroadcastSubscriptionsHandler)
	customer.GET("/subscriptions/unicast", handlers.GetCustomerUnicastSubscriptionsHandler)
	customer.GET("/vendors", handlers.GetCustomerSubscriptionsHandler)
}

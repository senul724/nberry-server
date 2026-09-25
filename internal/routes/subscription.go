package routes

import (
	"IAM-server/internal/handlers"
	"IAM-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetSubscriptionRoutes(api *gin.RouterGroup) {
	subs := api.Group("/subscriptions")
	subs.Use(middleware.ProtectRoute())

	// Broadcast subscriptions
	subs.POST("/broadcast", handlers.SubscribeBroadcastHandler)
	subs.DELETE("/broadcast/:app_id", handlers.UnsubscribeBroadcastHandler)

	// Unicast subscriptions
	subs.POST("/unicast", handlers.SubscribeUnicastHandler)
	subs.DELETE("/unicast/:app_id", handlers.UnsubscribeUnicastHandler)
}

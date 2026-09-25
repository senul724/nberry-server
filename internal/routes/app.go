package routes

import (
	"IAM-server/internal/handlers"
	"IAM-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetAppRoutes(api *gin.RouterGroup) {
	apps := api.Group("/apps")
	apps.Use(middleware.ProtectRoute())

	apps.GET("", handlers.GetAppsHandler)
	apps.GET("/:id", handlers.GetAppHandler)
	apps.POST("", handlers.CreateAppHandler)
	apps.DELETE("/:id", handlers.DeleteAppHandler)
}

package main

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/routes"
	"IAM-server/internal/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Notify Berry Server
// @version 1.0
// @description Backend API for Notify Berry server.
// @host localhost:3030
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	connections.ConnectDB()
	connections.ConnectRedis()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		connections.Migrate()
		return
	}

	r := gin.Default()

	// swagger docs
	r.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	// dynamic CORS for public and internal routes
	r.Use(utils.CORS())

	publicRoutes := r.Group("/api")
	internalRoutes := r.Group("/api")

	// auth routes
	routes.SetAuthRoutes(internalRoutes)

	// app routes
	routes.SetAppRoutes(internalRoutes)

	// subscription routes
	routes.SetSubscriptionRoutes(internalRoutes)

	// vendor routes (public push endpoints)
	routes.SetVendorRoutes(publicRoutes)

	r.Run(":3030")
}

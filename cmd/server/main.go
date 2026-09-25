package main

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/routes"
	"log"
	"time"

	"github.com/gin-contrib/cors"
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
	connections.Migrate()

	r := gin.Default()

	// swagger docs
	r.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	publicRoutes := r.Group("/api")
	internalRoutes := r.Group("/api")

	// CORS middleware internal
	internalRoutes.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:3002"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Device-Type", "X-Refresh-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// CORS middleware public
	publicRoutes.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // safe here, see below
		AllowCredentials: false,         // must be false if using "*"
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-API-Key"},
	}))

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

package routes

import (
	"fmt"
	"strings"

	adapter "github.com/bozoteam/roshan/src/database"
	"github.com/bozoteam/roshan/src/helpers"
	authControllers "github.com/bozoteam/roshan/src/modules/auth/controllers"
	"github.com/bozoteam/roshan/src/modules/auth/middlewares"
	authRouter "github.com/bozoteam/roshan/src/modules/auth/routes"
	chatControllers "github.com/bozoteam/roshan/src/modules/chat/controllers"
	chatRouter "github.com/bozoteam/roshan/src/modules/chat/routes"
	userRouter "github.com/bozoteam/roshan/src/modules/user/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes
func RegisterRoutes() *gin.Engine {
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	allowedOrigins := helpers.GetEnv("CORS_ALLOWED_ORIGINS")
	config.AllowOrigins = strings.Split(allowedOrigins, ",")
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	db := adapter.GetDBConnection()
	jwtConf := authControllers.NewJWTConfig()
	authController := authControllers.NewAuthController(db, jwtConf)
	authMiddleware := middlewares.NewAuthMiddleware(jwtConf)
	chatController := chatControllers.NewChatController()

	authRouter.RegisterAuthRoutes(router, authController)

	userRouter.RegisterUserRoutes(router, jwtConf, db)
	authRouter.RegisterAuthRoutes(router, authController)
	chatRouter.RegisterChatRoutes(router, authMiddleware, chatController)
	return router
}

// RunServer starts the API server
func RunServer() {
	port := helpers.GetEnv("API_PORT")
	if port == "" {
		fmt.Println("API_PORT is not set. Defaulting to 8080")
		port = "8080"
	}
	router := RegisterRoutes()
	router.Run(":" + port)
}

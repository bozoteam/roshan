package adapter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bozoteam/roshan/internal/helpers"
	"github.com/bozoteam/roshan/internal/modules/auth/middlewares"
	authRouter "github.com/bozoteam/roshan/internal/modules/auth/routes"
	authUsecase "github.com/bozoteam/roshan/internal/modules/auth/usecase"
	chatRouter "github.com/bozoteam/roshan/internal/modules/chat/routes"
	chatUsecase "github.com/bozoteam/roshan/internal/modules/chat/usecase"
	userRepository "github.com/bozoteam/roshan/internal/modules/user/repository"
	userRouter "github.com/bozoteam/roshan/internal/modules/user/routes"
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

	// router.SetTrustedProxies() TODO

	db := GetDBConnection()
	userRepository := userRepository.NewUserRepository(db)
	jwtConf := authUsecase.NewJWTConfig()
	authUsecase := authUsecase.NewAuthUsecase(userRepository, jwtConf)
	authMiddleware := middlewares.NewAuthMiddleware(jwtConf, userRepository)
	chatUsecase := chatUsecase.NewChatUsecase()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusText(http.StatusOK),
		})
	})

	authMiddlewareFunc := authMiddleware.AuthReqUser()

	userRouter.RegisterUserRoutes(router, jwtConf, db, authMiddlewareFunc)
	authRouter.RegisterAuthRoutes(router, authUsecase, authMiddlewareFunc)
	chatRouter.RegisterChatRoutes(router, authMiddleware, chatUsecase)
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

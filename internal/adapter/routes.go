package adapter

import (
	"fmt"
	"net/http"
	"strings"

	_ "github.com/bozoteam/roshan/docs"
	"github.com/bozoteam/roshan/internal/helpers"
	"github.com/bozoteam/roshan/internal/modules/auth/middlewares"
	jwtRepository "github.com/bozoteam/roshan/internal/modules/auth/repository/jwt"
	authRouter "github.com/bozoteam/roshan/internal/modules/auth/routes"
	authUsecase "github.com/bozoteam/roshan/internal/modules/auth/usecase"
	chatRouter "github.com/bozoteam/roshan/internal/modules/chat/routes"
	chatUsecase "github.com/bozoteam/roshan/internal/modules/chat/usecase"
	userRepository "github.com/bozoteam/roshan/internal/modules/user/repository"
	userRouter "github.com/bozoteam/roshan/internal/modules/user/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRoutes registers all routes
func RegisterRoutes() *gin.Engine {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
	jwtRepository := jwtRepository.NewJWTRepository()
	authUsecase := authUsecase.NewAuthUsecase(userRepository, jwtRepository)
	authMiddleware := middlewares.NewAuthMiddleware(jwtRepository, userRepository)
	chatUsecase := chatUsecase.NewChatUsecase(userRepository, jwtRepository)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusText(http.StatusOK),
		})
	})

	authMiddlewareFunc := authMiddleware.AuthReqUser()

	userRouter.RegisterUserRoutes(router, db, authMiddlewareFunc)
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

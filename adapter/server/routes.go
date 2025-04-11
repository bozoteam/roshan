package server

// import (
// 	"net/http"
// 	"strings"

// 	_ "github.com/bozoteam/roshan/docs"
// 	"github.com/bozoteam/roshan/helpers"
// 	"github.com/bozoteam/roshan/modules/auth/middlewares"
// 	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
// 	authRouter "github.com/bozoteam/roshan/modules/auth/routes"
// 	authUsecase "github.com/bozoteam/roshan/modules/auth/usecase"
// 	chatRouter "github.com/bozoteam/roshan/modules/chat/routes"
// 	chatUsecase "github.com/bozoteam/roshan/modules/chat/usecase"
// 	userRepository "github.com/bozoteam/roshan/modules/user/repository"
// 	userRouter "github.com/bozoteam/roshan/modules/user/routes"
// 	"github.com/gin-contrib/cors"
// 	"github.com/gin-gonic/gin"
// )

// // RegisterRoutes registers all routes
// func RegisterRoutes() *gin.Engine {
// 	router := gin.Default()

// 	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 	// Configure CORS
// 	config := cors.DefaultConfig()
// 	allowedOrigins := helpers.GetEnv("CORS_ALLOWED_ORIGINS")
// 	config.AllowOrigins = strings.Split(allowedOrigins, ",")
// 	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
// 	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
// 	config.AllowCredentials = true
// 	router.Use(cors.New(config))

// 	// router.SetTrustedProxies() TODO

// 	db := GetDBConnection()
// 	userRepository := userRepository.NewUserRepository(db)
// 	jwtRepository := jwtRepository.NewJWTRepository()
// 	authUsecase := authUsecase.NewAuthUsecase(userRepository, jwtRepository)
// 	authMiddleware := middlewares.NewAuthMiddleware(jwtRepository, userRepository)
// 	chatUsecase := chatUsecase.NewChatUsecase(userRepository, jwtRepository)

// 	router.GET("/health", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{
// 			"status": http.StatusText(http.StatusOK),
// 		})
// 	})

// 	authMiddlewareFunc := authMiddleware.AuthReqUser()

// 	userRouter.RegisterUserRoutes(router, db, authMiddlewareFunc)
// 	authRouter.RegisterAuthRoutes(router, authUsecase, authMiddlewareFunc)
// 	chatRouter.RegisterChatRoutes(router, authMiddleware, chatUsecase)
// 	return router
// }

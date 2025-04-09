package authRouter

import (
	authUsecase "github.com/bozoteam/roshan/internal/modules/auth/usecase"
	"github.com/gin-gonic/gin"
)

// registerAuthRoutes registers authentication routes
func RegisterAuthRoutes(router *gin.Engine, authUsecase *authUsecase.AuthUsecase, authMiddleware gin.HandlerFunc) {

	// Authentication routes
	router.POST("/auth", authUsecase.Authenticate)
	router.POST("/auth/refresh", authUsecase.Refresh)
}

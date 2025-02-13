package authRouter

import (
	"github.com/bozoteam/roshan/src/modules/auth/controllers"
	"github.com/gin-gonic/gin"
)

// registerAuthRoutes registers authentication routes
func RegisterAuthRoutes(router *gin.Engine, authController *controllers.AuthController) {

	// Authentication routes
	router.POST("/auth", authController.Authenticate)
	router.POST("/auth/refresh", authController.Refresh)
	router.GET("/auth/me", authController.GetLoggedInUser)
}

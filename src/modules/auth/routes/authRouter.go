package authRouter

import (
	"github.com/bozoteam/roshan/src/modules/auth/controllers"
	"github.com/gin-gonic/gin"
)

// registerAuthRoutes registers authentication routes
func RegisterAuthRoutes(router *gin.Engine) {

	// Authentication routes
	router.POST("/auth", controllers.Authenticate)
	router.POST("/auth/refresh", controllers.Refresh)
	router.GET("/auth/me", controllers.GetLoggedInUser)
}

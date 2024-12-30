package routes

import (
	"github.com/bozoteam/roshan/src/controllers"
	"github.com/gin-gonic/gin"
)

// registerAuthRoutes registers authentication routes
func registerAuthRoutes(router *gin.Engine) {

	// Authentication routes
	router.POST("/auth", controllers.Authenticate)
	router.POST("/refresh", controllers.Refresh)
}

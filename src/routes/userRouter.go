package routes

import (
	"github.com/bozoteam/roshan/src/controllers"
	"github.com/bozoteam/roshan/src/middlewares"
	"github.com/gin-gonic/gin"
)

// registerUserRoutes registers user related routes
func registerUserRoutes(router *gin.Engine) {

	// Public routes
	router.POST("/users", controllers.CreateUser)
	router.GET("/users/:username", controllers.FindUser)

	// Protected routes
	router.PUT("/users/:username", middlewares.AuthMiddleware(), controllers.UpdateUser)
	router.DELETE("/users/:username", middlewares.AuthMiddleware(), controllers.DeleteUser)
}

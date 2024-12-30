package userRouter

import (
	"github.com/bozoteam/roshan/src/modules/auth/middlewares"
	"github.com/bozoteam/roshan/src/modules/user/controllers"
	"github.com/gin-gonic/gin"
)

// registerUserRoutes registers user related routes
func RegisterUserRoutes(router *gin.Engine) {

	// Public routes
	router.POST("/users", controllers.CreateUser)
	router.GET("/users/:username", controllers.FindUser)

	// Protected routes
	router.PUT("/users/:username", middlewares.AuthMiddleware(), controllers.UpdateUser)
	router.DELETE("/users/:username", middlewares.AuthMiddleware(), controllers.DeleteUser)
}

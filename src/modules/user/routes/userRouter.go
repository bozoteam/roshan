package userRouter

import (
	authControllers "github.com/bozoteam/roshan/src/modules/auth/controllers"
	"github.com/bozoteam/roshan/src/modules/auth/middlewares"
	userControllers "github.com/bozoteam/roshan/src/modules/user/controllers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// registerUserRoutes registers user related routes
func RegisterUserRoutes(router *gin.Engine, jwtConf *authControllers.JWTConfig, db *gorm.DB) {

	userController := userControllers.NewUserController(db)

	// Public routes
	router.POST("/users", userController.CreateUser)
	router.GET("/users/:username", userController.FindUser)

	authMiddleware := middlewares.NewAuthMiddleware(jwtConf)
	// Protected routes
	router.PUT("/users/:username", authMiddleware.AuthReqUser(), userController.UpdateUser)
	router.DELETE("/users/:username", authMiddleware.AuthReqUser(), userController.DeleteUser)
}

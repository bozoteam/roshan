package userRouter

import (
	authControllers "github.com/bozoteam/roshan/src/modules/auth/controllers"
	userControllers "github.com/bozoteam/roshan/src/modules/user/controllers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// registerUserRoutes registers user related routes
func RegisterUserRoutes(router *gin.Engine, jwtConf *authControllers.JWTConfig, db *gorm.DB, authMiddleware gin.HandlerFunc) {
	userController := userControllers.NewUserController(db)

	// Public routes
	router.POST("/user", userController.CreateUser)
	// router.GET("/user/:username", userController.FindUser)

	// Protected routes
	router.PUT("/user", authMiddleware, userController.UpdateUser)
	router.DELETE("/user", authMiddleware, userController.DeleteUser)
	router.GET("/user", authMiddleware, userController.GetUser)
}

package userRouter

import (
	userUsecase "github.com/bozoteam/roshan/internal/modules/user/usecase"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// registerUserRoutes registers user related routes
func RegisterUserRoutes(router *gin.Engine, db *gorm.DB, authMiddleware gin.HandlerFunc) {
	userUsecase := userUsecase.NewUserUsecase(db)

	// Public routes
	router.POST("/user", userUsecase.CreateUser)
	// router.GET("/user/:username", userUsecase.FindUser)

	// Protected routes
	router.PUT("/user", authMiddleware, userUsecase.UpdateUser)
	router.DELETE("/user", authMiddleware, userUsecase.DeleteUser)
	router.GET("/user", authMiddleware, userUsecase.GetUser)
}

package chatRouter

import (
	"github.com/bozoteam/roshan/src/modules/chat/controllers"

	"github.com/bozoteam/roshan/src/modules/auth/middlewares"
	"github.com/gin-gonic/gin"
)

// RegisterChatRoutes registers chat related routes
func RegisterChatRoutes(router *gin.Engine) {

	// Public routes for chat
	router.POST("/chat/rooms", middlewares.AuthMiddleware(), controllers.CreateRoom)
	router.GET("/chat/rooms", middlewares.AuthMiddleware(), controllers.ListRooms)
	router.GET("/chat/rooms/:id/users", middlewares.AuthMiddleware(), controllers.ListUsers)
	router.DELETE("/chat/rooms/:id", middlewares.AuthMiddleware(), controllers.DeleteRoom)

	router.GET("/chat/rooms/:id/ws", controllers.HandleWebSocket)
}

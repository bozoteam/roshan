package chatRouter

import (
	"github.com/bozoteam/roshan/src/modules/chat/controllers"

	"github.com/bozoteam/roshan/src/modules/auth/middlewares"
	"github.com/gin-gonic/gin"
)

// RegisterChatRoutes registers chat related routes
func RegisterChatRoutes(router *gin.Engine, authMiddleware *middlewares.AuthMiddleware, chatController *controllers.ChatController) {
	authReqUser := authMiddleware.AuthReqUser()

	// Public routes for chat
	router.POST("/chat/rooms", authReqUser, chatController.CreateRoom)
	router.GET("/chat/rooms", authReqUser, chatController.ListRooms)
	router.GET("/chat/rooms/:id/users", authReqUser, chatController.ListUsers)
	router.DELETE("/chat/rooms/:id", authReqUser, chatController.DeleteRoom)

	router.GET("/chat/rooms/:id/ws", chatController.HandleWebSocket)
}

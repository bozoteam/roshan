package chatRouter

import (
	"github.com/bozoteam/roshan/internal/modules/chat/usecase"

	"github.com/bozoteam/roshan/internal/modules/auth/middlewares"
	"github.com/gin-gonic/gin"
)

// RegisterChatRoutes registers chat related routes
func RegisterChatRoutes(router *gin.Engine, authMiddleware *middlewares.AuthMiddleware, chatUsecase *usecase.ChatUsecase) {
	authReqUser := authMiddleware.AuthReqUser()

	// Public routes for chat
	router.GET("/chat/rooms", chatUsecase.ListRooms)

	// private
	router.POST("/chat/rooms", authReqUser, chatUsecase.CreateRoom)

	// router.GET("/chat/rooms/:id/users", authReqUser, chatUsecase.ListUsers)
	router.DELETE("/chat/rooms/:id", authReqUser, chatUsecase.DeleteRoom)

	router.GET("/chat/rooms/:id/ws", authReqUser, chatUsecase.HandleWebSocket)
}

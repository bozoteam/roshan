package usecase

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	log "github.com/bozoteam/roshan/internal/adapter/log"
	"github.com/bozoteam/roshan/internal/modules/chat/models"
	userModel "github.com/bozoteam/roshan/internal/modules/user/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ChatUsecase struct {
	hub      *models.Hub
	upgrader *websocket.Upgrader
	logger   *slog.Logger
}

func NewChatUsecase() *ChatUsecase {
	hub := models.NewHub()
	go hub.Run()
	return &ChatUsecase{
		hub:      hub,
		upgrader: new(websocket.Upgrader),
		logger:   log.LogWithModule("chat_usecase"),
	}
}

// RoomResponse represents a chat room with its users
type RoomResponse struct {
	Id      string   `json:"id" example:"f81d4fae-7dec-11d0-a765-00a0c91e6bf6"`
	Name    string   `json:"name" example:"General Discussion"`
	UserIds []string `json:"users" example:"00DBA43A-5F6D-46DE-A07C-E76FB55435ED,DFD4FC58-6FB7-4CDC-97F6-151E4125B617"`
}

// RoomCreateRequest represents data needed to create a chat room
type RoomCreateRequest struct {
	RoomName string `json:"room_name" binding:"required" example:"General Discussion"`
}

// RoomCreateResponse represents the response after creating a room
type RoomCreateResponse struct {
	Id string `json:"id" example:"f81d4fae-7dec-11d0-a765-00a0c91e6bf6"`
}

// MessageRequest represents the structure of a chat message
type MessageRequest struct {
	Message string `json:"message" binding:"required" example:"Hello, everyone!"`
}

// UserInfo represents basic user information for room membership
type UserInfo struct {
	ID       string `json:"id" example:"123"`
	Username string `json:"username" example:"john_doe"`
}

// Helper function to send user list update event
func (cc *ChatUsecase) sendUserListUpdate(room *models.Room) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	users := make([]*userModel.User, 0, len(room.Clients))
	for client := range room.Clients {
		users = append(users, client.User)
	}

	// Create a user list event
	event := &models.Event{
		RoomID:    room.ID,
		Users:     users,
		Timestamp: time.Now().Unix(),
	}

	// Broadcast the event
	cc.hub.BrodcastEvent <- event
}

// SendMessage godoc
// @Summary Send a message to a chat room
// @Description Send a message to a specific chat room that the user is a member of
// @Tags chat
// @Accept json
// @Produce json
// @Param id path string true "Room ID"
// @Param message body MessageRequest true "Message content"
// @Security BearerAuth
// @Success 200 {object} nil
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 403 {object} map[string]string "User not in room"
// @Failure 404 {object} map[string]string "Room not found"
// @Router /chat/rooms/message/{id} [post]
func (cc *ChatUsecase) SendMessage(context *gin.Context) {
	user := context.MustGet("user").(*userModel.User)
	roomID := context.Param("id")

	cc.hub.Mutex.Lock()
	room, exists := cc.hub.Rooms[roomID]
	cc.hub.Mutex.Unlock()

	if !exists {
		context.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	foundUser := false
	room.Mutex.Lock()
	for client := range room.Clients {
		if client.User.Id == user.Id {
			foundUser = true
			break
		}
	}
	room.Mutex.Unlock()

	if !foundUser {
		context.JSON(http.StatusForbidden, gin.H{"error": "User not found in room"})
		return
	}

	var input MessageRequest
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Create message with proper metadata
	message := &models.Message{
		RoomID:    roomID,
		UserID:    user.Id,
		Content:   input.Message,
		Timestamp: time.Now().Unix(),
	}

	// Broadcast the message
	cc.hub.BroadcastMessage <- message
	context.JSON(http.StatusOK, nil)
}

// CreateRoom godoc
// @Summary Create a new chat room
// @Description Create a new chat room with the authenticated user as creator
// @Tags chat
// @Accept json
// @Produce json
// @Param room body RoomCreateRequest true "Room information"
// @Security BearerAuth
// @Success 200 {object} RoomCreateResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /chat/rooms [post]
func (cc *ChatUsecase) CreateRoom(context *gin.Context) {
	user := context.MustGet("user").(*userModel.User)

	var input RoomCreateRequest
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		cc.logger.Error("Failed to generate UUID", "error", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	room := models.Room{
		ID:        uuid.String(),
		CreatorId: user.Id,
		Name:      input.RoomName,
		Clients:   make(map[*models.RoomClient]bool),
		Mutex:     &sync.Mutex{},
	}

	cc.hub.Register <- &room
	context.JSON(http.StatusOK, RoomCreateResponse{Id: room.ID})
}

// ListRooms godoc
// @Summary List all chat rooms
// @Description Get a list of all available chat rooms and their members
// @Tags chat
// @Produce json
// @Success 200 {array} RoomResponse
// @Router /chat/rooms [get]
func (cc *ChatUsecase) ListRooms(context *gin.Context) {
	cc.hub.Mutex.Lock()
	defer cc.hub.Mutex.Unlock()

	rooms := make([]RoomResponse, 0, len(cc.hub.Rooms))
	for _, room := range cc.hub.Rooms {
		room.Mutex.Lock()
		users := make([]string, 0, len(room.Clients))
		for client := range room.Clients {
			users = append(users, client.Id)
		}

		roomResponse := RoomResponse{
			Id:      room.ID,
			Name:    room.Name,
			UserIds: users,
		}
		room.Mutex.Unlock()

		rooms = append(rooms, roomResponse)
	}

	context.JSON(http.StatusOK, rooms)
}

// DeleteRoom godoc
// @Summary Delete a chat room
// @Description Delete a specific chat room
// @Tags chat
// @Produce json
// @Param id path string true "Room ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string "Room not found"
// @Router /chat/rooms/{id} [delete]
func (cc *ChatUsecase) DeleteRoom(c *gin.Context) {
	roomID := c.Param("id")

	cc.hub.Mutex.Lock()
	room, exists := cc.hub.Rooms[roomID]
	cc.hub.Mutex.Unlock()

	if exists {
		cc.hub.Unregister <- room
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
	}
}

// HandleWebSocket godoc
// @Summary Connect to a chat room via WebSocket
// @Description Establish a WebSocket connection to a specific chat room
// @Tags chat
// @Param id path string true "Room ID"
// @Security BearerAuth
// @Failure 404 {object} map[string]string "Room not found"
// @Router /chat/rooms/{id}/ws [get]
func (cc *ChatUsecase) HandleWebSocket(context *gin.Context) {
	user := context.MustGet("user").(*userModel.User)
	roomID := context.Param("id")

	cc.hub.Mutex.Lock()
	room, exists := cc.hub.Rooms[roomID]
	cc.hub.Mutex.Unlock()

	if !exists {
		context.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Allow all origins for the WebSocket upgrade
	cc.upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := cc.upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		cc.logger.Error("Failed to upgrade connection", "error", err)
		return
	}

	roomUser := &models.RoomClient{
		User: user,
		Conn: conn,
	}

	// Add user to the room
	room.Mutex.Lock()
	room.Clients[roomUser] = true
	room.Mutex.Unlock()

	// Send user list update after adding user
	cc.sendUserListUpdate(room)

	cc.logger.Info("User connected to room", "user_id", user.Id, "room_id", roomID)

	defer func() {
		conn.Close()

		// Remove user from the room
		room.Mutex.Lock()
		delete(room.Clients, roomUser)
		room.Mutex.Unlock()

		cc.logger.Info("User disconnected from room", "user_id", user.Id, "room_id", roomID)

		// Send updated user list after removing user
		cc.sendUserListUpdate(room)
	}()

	// Keep connection alive but don't process messages
	for {
		// Only handle ping/pong or connection close events
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// Don't process or broadcast messages from websocket
	}
}

package usecase

import (
	"net/http"
	"sync"

	"github.com/bozoteam/roshan/internal/modules/chat/models"
	userModel "github.com/bozoteam/roshan/internal/modules/user/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ChatUsecase struct {
	hub      *models.Hub
	upgrader *websocket.Upgrader
}

func NewChatUsecase() *ChatUsecase {
	hub := models.NewHub()
	go hub.Run()
	return &ChatUsecase{hub: hub, upgrader: new(websocket.Upgrader)}
}

type RoomResponse struct {
	Id      string   `json:"id"`
	UserIds []string `json:"users"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (cc *ChatUsecase) CreateRoom(context *gin.Context) {
	user := context.MustGet("user").(*userModel.User)

	var input struct {
		RoomName string `json:"room_name" binding:"required"`
	}
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	room := models.Room{
		ID:        uuid.String(),
		CreatorId: user.Id,
		Name:      input.RoomName,
		Clients:   make(map[*models.RoomClient]bool),
		Mutex:     &sync.Mutex{},
	}
	cc.hub.Register <- &room

	context.JSON(http.StatusOK, gin.H{"id": room.ID})
}

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
			UserIds: users,
		}
		room.Mutex.Unlock()
		rooms = append(rooms, roomResponse)
	}
	context.JSON(http.StatusOK, rooms)
}

// func (cc *ChatUsecase) ListUsers(c *gin.Context) {
// 	roomID := c.Param("id")

// 	cc.hub.Mutex.Lock()
// 	room, exists := cc.hub.Rooms[roomID]
// 	cc.hub.Mutex.Unlock()
// 	if !exists {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
// 		return
// 	}
// 	room.Mutex.Lock()
// 	users := make([]UserInfo, 0, len(room.Clients))
// 	for _, user := range room.Clients {
// 		users = append(users, UserInfo{ID: user.ID, Username: user.Username})
// 	}
// 	room.Mutex.Unlock()
// 	c.JSON(http.StatusOK, users)
// }

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

	conn, err := cc.upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		return
	}

	roomUser := models.RoomClient{
		User: user,
		Conn: conn,
	}

	room.Mutex.Lock()
	room.Clients[&roomUser] = true
	room.Mutex.Unlock()

	defer func() {
		conn.Close()
		room.Mutex.Lock()
		delete(room.Clients, &roomUser)
		room.Mutex.Unlock()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		cc.hub.Broadcast <- models.Message{RoomID: roomID, Content: string(msg)}
	}
}

package models

// roomRequest is used to safely get room data
type roomRequest struct {
	id string

	result chan *Room
}

// roomsRequest is used to get a list of all rooms
type roomsRequest struct {
	result chan []*Room
}

type deleteRoom struct {
	roomId string

	result chan *Room
}

type sendMessage struct {
	Msg *Message

	result chan *Message
}

type createRoom struct {
	Room *Room

	result chan *Room
}

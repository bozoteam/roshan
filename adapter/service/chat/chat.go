package chat_service

import (
	"context"

	gen "github.com/bozoteam/roshan/adapter/grpc/gen/chat"
	userGen "github.com/bozoteam/roshan/adapter/grpc/gen/user"
	"github.com/bozoteam/roshan/modules/chat/usecase"
)

type ChatService struct {
	chatUsecase *usecase.ChatUsecase
	gen.UnimplementedChatServiceServer
}

func NewChatService(chatUsecase *usecase.ChatUsecase) *ChatService {
	return &ChatService{
		chatUsecase: chatUsecase,
	}
}

func (s *ChatService) SendMessage(ctx context.Context, req *gen.SendMessageRequest) (*gen.SendMessageResponse, error) {
	err := s.chatUsecase.SendMessage(ctx, req.Content, req.RoomId)
	if err != nil {
		return nil, err
	}

	return &gen.SendMessageResponse{}, nil
}

func (s *ChatService) CreateRoom(ctx context.Context, req *gen.CreateRoomRequest) (*gen.CreateRoomResponse, error) {
	room, err := s.chatUsecase.CreateRoom(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &gen.CreateRoomResponse{
		Room: &gen.Room{
			Name:      room.Name,
			Id:        room.ID,
			CreatorId: room.CreatorID,
			Users:     nil,
		},
	}, nil
}

func (s *ChatService) ListRooms(ctx context.Context, req *gen.ListRoomsRequest) (*gen.ListRoomsResponse, error) {
	rooms, err := s.chatUsecase.ListRooms(ctx)
	if err != nil {
		return nil, err
	}

	outRooms := make([]*gen.Room, 0, len(rooms))
	for _, room := range rooms {
		users := make([]*userGen.User, 0, len(room.Users))
		for _, user := range room.Users {
			users = append(users, &userGen.User{
				Name:  user.Name,
				Id:    user.Id,
				Email: user.Email,
			})
		}

		outRooms = append(outRooms, &gen.Room{
			Id:        room.Id,
			CreatorId: room.CreatorId,
			Users:     users,
			Name:      room.Name,
		})

	}

	return &gen.ListRoomsResponse{
		Rooms: outRooms,
	}, nil
}

func (s *ChatService) DeleteRoom(ctx context.Context, req *gen.DeleteRoomRequest) (*gen.DeleteRoomResponse, error) {
	room, err := s.chatUsecase.DeleteRoom(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	users := make([]*userGen.User, 0, len(room.Users))
	for _, user := range room.Users {
		users = append(users, &userGen.User{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Name,
		})
	}

	return &gen.DeleteRoomResponse{
		Room: &gen.Room{
			Id:        room.Id,
			CreatorId: room.CreatorId,
			Users:     users,
		},
	}, nil
}

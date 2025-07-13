package chat_service

import (
	"context"

	gen "github.com/bozoteam/roshan/adapter/grpc/gen/chat"
	commonGen "github.com/bozoteam/roshan/adapter/grpc/gen/common"
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
		Room: room.ToGRPCRoom(),
	}, nil
}

func (s *ChatService) ListRooms(ctx context.Context, req *gen.ListRoomsRequest) (*gen.ListRoomsResponse, error) {
	rooms, err := s.chatUsecase.ListRooms(ctx)
	if err != nil {
		return nil, err
	}

	outRooms := make([]*commonGen.Room, 0, len(rooms))
	for _, room := range rooms {
		outRooms = append(outRooms, room.ToGRPCRoom())
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

	return &gen.DeleteRoomResponse{
		Room: room.ToGRPCRoom(),
	}, nil
}

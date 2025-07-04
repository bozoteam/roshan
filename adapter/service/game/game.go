package game_service

import (
	"context"

	commonGen "github.com/bozoteam/roshan/adapter/grpc/gen/common"
	gen "github.com/bozoteam/roshan/adapter/grpc/gen/game"
	"github.com/bozoteam/roshan/modules/game/usecase"
)

type GameService struct {
	chatUsecase *usecase.GameUsecase
	gen.UnimplementedGameServiceServer
}

func NewGameService(gameUsecase *usecase.GameUsecase) *GameService {
	return &GameService{
		chatUsecase: gameUsecase,
	}
}

func (s *GameService) CreateRoom(ctx context.Context, req *gen.CreateGameRoomRequest) (*gen.CreateGameRoomResponse, error) {
	room, err := s.chatUsecase.CreateRoom(ctx, req.Name, req.GetName())
	if err != nil {
		return nil, err
	}

	return &gen.CreateGameRoomResponse{
		Room: &commonGen.Room{
			Name:      room.Name,
			Id:        room.ID,
			CreatorId: room.CreatorID,
			Users:     nil,
		},
	}, nil
}

// func (s *GameService) ListRooms(ctx context.Context, req *gen.ListRoomsRequest) (*gen.ListRoomsResponse, error) {
// 	rooms, err := s.chatUsecase.ListRooms(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	outRooms := make([]*commonGen.Room, 0, len(rooms))
// 	for _, room := range rooms {
// 		users := make([]*userGen.User, 0, len(room.Users))
// 		for _, user := range room.Users {
// 			users = append(users, &userGen.User{
// 				Name:  user.Name,
// 				Id:    user.Id,
// 				Email: user.Email,
// 			})
// 		}

// 		outRooms = append(outRooms, &commonGen.Room{
// 			Id:        room.Id,
// 			CreatorId: room.CreatorId,
// 			Users:     users,
// 			Name:      room.Name,
// 		})

// 	}

// 	return &gen.ListRoomsResponse{
// 		Rooms: outRooms,
// 	}, nil
// }

// func (s *GameService) DeleteRoom(ctx context.Context, req *gen.DeleteRoomRequest) (*gen.DeleteRoomResponse, error) {
// 	room, err := s.chatUsecase.DeleteRoom(ctx, req.Id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	users := make([]*userGen.User, 0, len(room.Users))
// 	for _, user := range room.Users {
// 		users = append(users, &userGen.User{
// 			Id:    user.Id,
// 			Name:  user.Name,
// 			Email: user.Name,
// 		})
// 	}

// 	return &gen.DeleteRoomResponse{
// 		Room: &commonGen.Room{
// 			Id:        room.Id,
// 			CreatorId: room.CreatorId,
// 			Users:     users,
// 		},
// 	}, nil
// }

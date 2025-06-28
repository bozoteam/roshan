package usecase

import (
	"context"
	"log/slog"

	"github.com/bozoteam/roshan/adapter/log"
	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
	chatModel "github.com/bozoteam/roshan/modules/chat/models"
	gameModel "github.com/bozoteam/roshan/modules/game/models"
	userModel "github.com/bozoteam/roshan/modules/user/models"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
	ws_hub "github.com/bozoteam/roshan/modules/websocket/hub"
)

type GameUsecase struct {
	hub            *ws_hub.Hub
	logger         *slog.Logger
	jwtRepository  *jwtRepository.JWTRepository
	userRepository *userRepository.UserRepository
}

func NewGameUsecase(
	jwtRepository *jwtRepository.JWTRepository,
	userRepository *userRepository.UserRepository,
) *GameUsecase {
	return &GameUsecase{
		hub:            ws_hub.NewHub(),
		logger:         log.LogWithModule("game_usecase"),
		jwtRepository:  jwtRepository,
		userRepository: userRepository,
	}
}

func (u *GameUsecase) CreateRoom(ctx context.Context, name string) *chatModel.Room {
	user := ctx.Value("user").(*userModel.User)

	room := gameModel.NewGameRoom(name, user.Id)

	u.hub.CreateRoom(room)

	return nil
}

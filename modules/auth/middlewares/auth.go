package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/bozoteam/roshan/adapter/log"
	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthMiddleware struct {
	logger             *slog.Logger
	jwtRepository      *jwtRepository.JWTRepository
	userRepository     *userRepository.UserRepository
	blacklistedMethods map[string]struct{}
}

func NewAuthMiddleware(jwtRepository *jwtRepository.JWTRepository, userRepository *userRepository.UserRepository, blacklistedMethods map[string]struct{}) *AuthMiddleware {
	return &AuthMiddleware{jwtRepository: jwtRepository, userRepository: userRepository, logger: log.LogWithModule("auth_middleware"), blacklistedMethods: blacklistedMethods}
}

func (m *AuthMiddleware) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if _, ok := m.blacklistedMethods[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing token")
	}

	authorization, ok := md["authorization"]
	if !ok {
		return nil, errors.New("missing token")
	}

	if len(authorization) != 1 {
		return nil, errors.New("wrong token format")
	}

	_token := strings.Split(authorization[0], " ")
	if len(_token) != 2 {
		return nil, errors.New("wrong token format")
	}
	token := _token[1]

	_, claims, err := m.jwtRepository.ValidateToken(token, jwtRepository.ACCESS_TOKEN)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return nil, errors.New("invalid token")
	}

	user, err := m.userRepository.FindUserById(subject)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	ctx = context.WithValue(ctx, "user", user)

	return handler(ctx, req)
}

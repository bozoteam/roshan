package auth_service

import (
	"context"

	gen "github.com/bozoteam/roshan/adapter/grpc/gen/auth"
	"github.com/bozoteam/roshan/modules/auth/usecase"
)

type AuthService struct {
	authUsecase *usecase.AuthUsecase
	gen.UnimplementedAuthServiceServer
}

func NewAuthService(authUsecase *usecase.AuthUsecase) *AuthService {
	return &AuthService{
		authUsecase: authUsecase,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, req *gen.AuthenticateRequest) (*gen.AuthenticateResponse, error) {
	tokenData, err := s.authUsecase.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &gen.AuthenticateResponse{
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
		TokenType:    tokenData.TokenType,
		ExpiresIn:    uint64(tokenData.ExpiresIn),
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *gen.RefreshTokenRequest) (*gen.RefreshTokenResponse, error) {
	tokenData, err := s.authUsecase.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &gen.RefreshTokenResponse{
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
		TokenType:    tokenData.TokenType,
		ExpiresIn:    uint64(tokenData.ExpiresIn),
	}, nil
}

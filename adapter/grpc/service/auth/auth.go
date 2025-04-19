package auth_service

import (
	"context"
	"net/http"

	gen "github.com/bozoteam/roshan/adapter/grpc/gen/auth"
	"github.com/bozoteam/roshan/modules/auth/usecase"
	"github.com/bozoteam/roshan/roshan_errors"
	"google.golang.org/grpc/metadata"
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
	var inputRefreshToken string

	if req.RefreshToken != nil {
		inputRefreshToken = *req.RefreshToken
	} else {
		// get from cookie not sure TODO: check
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, roshan_errors.ErrInvalidInput
		}

		cookieValues := md.Get("cookie")
		if len(cookieValues) == 0 {
			return nil, roshan_errors.ErrInvalidInput
		}

		header := http.Header{}
		header.Add("Cookie", cookieValues[0])
		request := http.Request{Header: header}

		cookie, err := request.Cookie("refresh_token")
		if err != nil || cookie.Value == "" {
			return nil, roshan_errors.ErrInvalidInput
		}

		inputRefreshToken = cookie.Value
	}

	tokenData, err := s.authUsecase.Refresh(ctx, inputRefreshToken)
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

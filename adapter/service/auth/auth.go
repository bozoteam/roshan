package auth_service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	gen "github.com/bozoteam/roshan/adapter/grpc/gen/auth"
	"github.com/bozoteam/roshan/adapter/log"
	"github.com/bozoteam/roshan/modules/auth/usecase"
	"github.com/bozoteam/roshan/roshan_errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthService struct {
	logger      *slog.Logger
	authUsecase *usecase.AuthUsecase
	gen.UnimplementedAuthServiceServer
}

func NewAuthService(authUsecase *usecase.AuthUsecase) *AuthService {
	return &AuthService{
		authUsecase: authUsecase,
		logger:      log.LogWithModule("auth_service"),
	}
}

func genAuthResponseFromToken(token *usecase.TokenResponse) *gen.AuthenticateResponse {
	return &gen.AuthenticateResponse{
		RefreshExpiresIn: token.RefreshExpiresIn,
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		TokenType:        token.TokenType,
		ExpiresIn:        token.ExpiresIn,
	}
}

func (s *AuthService) setAuthCookie(ctx context.Context, respToken *usecase.TokenResponse) {
	md := metadata.Pairs(
		"Set-Cookie", fmt.Sprintf("access_token=%s; HttpOnly; SameSite=Strict; Path=/api; Max-Age=%d",
			respToken.AccessToken,
			respToken.ExpiresIn,
		),
		"Set-Cookie", fmt.Sprintf("refresh_token=%s; HttpOnly; SameSite=Strict; Path=/api/v1/auth/refresh; Max-Age=%d",
			respToken.RefreshToken,
			respToken.RefreshExpiresIn,
		),
		"Cache-Control", "no-store",
	)
	if err := grpc.SetHeader(ctx, md); err != nil {
		s.logger.Error("Failed to set cookie header", "error", err)
	}
}

func (s *AuthService) deleteAuthCookie(ctx context.Context) {
	md := metadata.Pairs(
		"Set-Cookie", fmt.Sprintf("access_token=deleted; HttpOnly; SameSite=Strict; Path=/api; Max-Age=0"),
		"Set-Cookie", fmt.Sprintf("refresh_token=deleted; HttpOnly; SameSite=Strict; Path=/api/v1/auth/refresh; Max-Age=0"),
		"Cache-Control", "no-store",
	)
	if err := grpc.SetHeader(ctx, md); err != nil {
		s.logger.Error("Failed to set cookie header", "error", err)
	}
}

func (s *AuthService) Authenticate(ctx context.Context, req *gen.AuthenticateRequest) (*gen.AuthenticateResponse, error) {
	tokenData, err := s.authUsecase.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	s.setAuthCookie(ctx, tokenData)

	return genAuthResponseFromToken(tokenData), nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *gen.RefreshTokenRequest) (*gen.AuthenticateResponse, error) {
	var inputRefreshToken string

	if req.RefreshToken != nil {
		inputRefreshToken = *req.RefreshToken
	} else {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, roshan_errors.ErrInvalidRequest
		}

		cookieValues := md.Get("cookie")
		if len(cookieValues) != 1 {
			return nil, roshan_errors.ErrInvalidRequest
		}

		cookieStr := cookieValues[0]

		cookies, err := http.ParseCookie(cookieStr)
		if err != nil {
			return nil, roshan_errors.ErrInvalidRequest
		}

		var found bool
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				found = true
				inputRefreshToken = cookie.Value
				break
			}
		}

		if !found {
			return nil, roshan_errors.ErrInvalidRequest
		}
	}

	tokenData, err := s.authUsecase.Refresh(ctx, inputRefreshToken)
	if err != nil {
		return nil, err
	}

	s.setAuthCookie(ctx, tokenData)

	return genAuthResponseFromToken(tokenData), nil
}

func (s *AuthService) Logout(ctx context.Context, req *gen.LogoutRequest) (*gen.LogoutResponse, error) {
	s.authUsecase.Logout(ctx)

	s.deleteAuthCookie(ctx)

	return &gen.LogoutResponse{}, nil
}

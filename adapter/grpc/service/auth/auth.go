package auth_service

import (
	"context"
	"fmt"
	"net/http"

	gen "github.com/bozoteam/roshan/adapter/grpc/gen/auth"
	"github.com/bozoteam/roshan/modules/auth/usecase"
	"github.com/bozoteam/roshan/roshan_errors"
	"google.golang.org/grpc"
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

func genAuthResponseFromToken(token *usecase.TokenResponse) *gen.AuthenticateResponse {
	return &gen.AuthenticateResponse{
		RefreshExpiresIn: token.RefreshExpiresIn,
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		TokenType:        token.TokenType,
		ExpiresIn:        token.ExpiresIn,
	}
}

func (s *AuthService) setRefreshTokenCookie(ctx context.Context, token string, expiration uint64) {
	md := metadata.Pairs(
		"Set-Cookie", fmt.Sprintf("refresh_token=%s; HttpOnly; SameSite=Strict; Path=/api; Max-Age=%d",
			token,
			expiration,
		),
		"Cache-Control", "no-store",
	)
	if err := grpc.SetHeader(ctx, md); err != nil {
		// Log error but don't fail the request
		// log.Printf("Failed to set cookie header: %v", err)
	}

}

func (s *AuthService) deleteRefreshTokenCookie(ctx context.Context) {
	md := metadata.Pairs(
		"Set-Cookie", fmt.Sprintf("refresh_token=\"\"; HttpOnly; SameSite=Strict; Path=/api; Max-Age=0"),
		"Cache-Control", "no-store",
	)
	if err := grpc.SetHeader(ctx, md); err != nil {
		// Log error but don't fail the request
		// log.Printf("Failed to set cookie header: %v", err)
	}

}

func (s *AuthService) Authenticate(ctx context.Context, req *gen.AuthenticateRequest) (*gen.AuthenticateResponse, error) {
	tokenData, err := s.authUsecase.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	s.setRefreshTokenCookie(ctx, tokenData.RefreshToken, tokenData.RefreshExpiresIn)

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

	s.setRefreshTokenCookie(ctx, tokenData.RefreshToken, tokenData.RefreshExpiresIn)

	return genAuthResponseFromToken(tokenData), nil
}

func (s *AuthService) Logout(ctx context.Context, req *gen.LogoutRequest) (*gen.LogoutResponse, error) {
	s.authUsecase.Logout(ctx)

	s.deleteRefreshTokenCookie(ctx)

	return &gen.LogoutResponse{}, nil
}

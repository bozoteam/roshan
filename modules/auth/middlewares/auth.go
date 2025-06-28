package middlewares

import (
	"context"
	"log/slog"
	"strings"

	"github.com/bozoteam/roshan/adapter/log"
	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
	"github.com/bozoteam/roshan/roshan_errors"
	"github.com/gin-gonic/gin"
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

func authErrorLikeGRPC(err error) gin.H {
	return gin.H{
		"code":    16,
		"message": err.Error(),
		"details": []any{},
	}
}

func (m *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var token string
		tokenHeader := ctx.GetHeader("authorization")

		// First try authorization header
		if tokenHeader != "" {
			tokenParts := strings.Split(tokenHeader, " ")
			if len(tokenParts) != 2 {
				m.logger.Error("invalid authorization header format")
				ctx.AbortWithStatusJSON(401, authErrorLikeGRPC(roshan_errors.ErrWrongTokenFormat))
				return
			}
			token = tokenParts[1]
		} else {
			// Fallback to cookie
			cookie, err := ctx.Cookie("access_token")
			if err != nil {
				m.logger.Error("missing authorization header and cookie")
				ctx.AbortWithStatusJSON(401, authErrorLikeGRPC(roshan_errors.ErrMissingToken))
				return
			}
			token = cookie
		}

		// Validate token and proceed
		_, claims, err := m.jwtRepository.ValidateToken(token, jwtRepository.ACCESS_TOKEN)
		if err != nil {
			ctx.AbortWithStatusJSON(401, authErrorLikeGRPC(roshan_errors.ErrInvalidToken))
			return
		}

		subject, err := claims.GetSubject()
		if err != nil {
			ctx.AbortWithStatusJSON(401, authErrorLikeGRPC(roshan_errors.ErrInvalidToken))
			return
		}

		user, err := m.userRepository.FindUserById(subject)
		if err != nil {
			ctx.AbortWithStatusJSON(401, authErrorLikeGRPC(roshan_errors.ErrInvalidToken))
			return
		}

		ctx.Set("user", user)
		ctx.Next()
	}
}

func (m *AuthMiddleware) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if _, ok := m.blacklistedMethods[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		m.logger.Error("missing metadata")
		return nil, roshan_errors.ErrMissingToken
	}

	var token string

	// First try authorization header
	authorization, ok := md["authorization"]
	if ok {
		if len(authorization) != 1 {
			m.logger.Error("invalid authorization header format")
			return nil, roshan_errors.ErrWrongTokenFormat
		}

		tokenParts := strings.Split(authorization[0], " ")
		if len(tokenParts) == 2 {
			token = tokenParts[1]
		} else {
			return nil, roshan_errors.ErrWrongTokenFormat
		}
	} else {
		// Fallback to cookie
		cookies, hasCookies := md["cookie"]
		if !hasCookies {
			m.logger.Error("missing authorization header and cookie")
			return nil, roshan_errors.ErrMissingToken
		}

		// Parse access_token from cookies
		for _, c := range cookies {
			if strings.HasPrefix(c, "access_token=") {
				token = strings.TrimPrefix(c, "access_token=")
				break
			}
		}

		if token == "" {
			m.logger.Error("access_token cookie not found")
			return nil, roshan_errors.ErrMissingToken
		}
	}

	// Validate token and proceed
	_, claims, err := m.jwtRepository.ValidateToken(token, jwtRepository.ACCESS_TOKEN)
	if err != nil {
		return nil, roshan_errors.ErrInvalidToken
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return nil, roshan_errors.ErrInvalidToken
	}

	user, err := m.userRepository.FindUserById(subject)
	if err != nil {
		return nil, roshan_errors.ErrInvalidToken
	}

	ctx = context.WithValue(ctx, "user", user)
	return handler(ctx, req)
}

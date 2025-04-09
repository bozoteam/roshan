package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bozoteam/roshan/internal/adapter/log"
	jwtRepository "github.com/bozoteam/roshan/internal/modules/auth/repository/jwt"
	userRepository "github.com/bozoteam/roshan/internal/modules/user/repository"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	logger         *slog.Logger
	jwtRepository  *jwtRepository.JWTRepository
	userRepository *userRepository.UserRepository
}

func NewAuthMiddleware(jwtRepository *jwtRepository.JWTRepository, userRepository *userRepository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{jwtRepository: jwtRepository, userRepository: userRepository, logger: log.LogWithModule("auth_middleware")}
}

func (m *AuthMiddleware) AuthReqUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		authorizationHeader := context.GetHeader("Authorization")

		tokenStringList := strings.Split(authorizationHeader, "Bearer ")
		if len(tokenStringList) != 2 {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			context.Abort()
			return
		}

		tokenString := tokenStringList[1]

		if tokenString == "" {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			context.Abort()
			return
		}

		_, claims, err := m.jwtRepository.ValidateToken(tokenString, jwtRepository.ACCESS_TOKEN)
		if err != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			context.Abort()
			return
		}

		subject, err := claims.GetSubject()
		if err != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			context.Abort()
			return
		}

		user, err := m.userRepository.FindUserById(subject)
		if err != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			context.Abort()
			return
		}

		context.Set("user", user)
		context.Next()
	}
}

package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	log "github.com/bozoteam/roshan/internal/log"
	authUsecase "github.com/bozoteam/roshan/internal/modules/auth/usecase"
	userRepository "github.com/bozoteam/roshan/internal/modules/user/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	logger         *slog.Logger
	jwtConfig      *authUsecase.JWTConfig
	userRepository *userRepository.UserRepository
}

func NewAuthMiddleware(jwtConf *authUsecase.JWTConfig, userRepository *userRepository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{jwtConfig: jwtConf, userRepository: userRepository, logger: log.WithModule("auth_middleware")}
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

		token, err := jwt.Parse(tokenString, m.jwtConfig.GetTokenKeyFunc)
		if err != nil || !token.Valid {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			context.Abort()
			return
		}

		subject, err := token.Claims.GetSubject()
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

		context.Set("user", &user)
		context.Next()
	}
}

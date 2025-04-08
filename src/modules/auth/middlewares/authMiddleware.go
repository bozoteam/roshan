package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	log "github.com/bozoteam/roshan/src/log"
	"github.com/bozoteam/roshan/src/modules/auth/controllers"
	models "github.com/bozoteam/roshan/src/modules/user/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthMiddleware struct {
	logger    *slog.Logger
	jwtConfig *controllers.JWTConfig
	db        *gorm.DB
}

func NewAuthMiddleware(jwtConf *controllers.JWTConfig, db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{jwtConfig: jwtConf, db: db, logger: log.WithModule("auth_middleware")}
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

		var user models.User
		err = m.db.First(&user, "id = ?", subject).Error
		if err != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			context.Abort()
			return
		}

		context.Set("user", &user)
		context.Next()
	}
}

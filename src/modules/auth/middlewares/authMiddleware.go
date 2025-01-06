package middlewares

import (
	"net/http"

	"github.com/bozoteam/roshan/src/modules/auth/controllers"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtConfig *controllers.JWTConfig
}

func NewAuthMiddleware(jwtConf *controllers.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{jwtConfig: jwtConf}
}

func (m *AuthMiddleware) AuthReqUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
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

		context.Set("username", subject)
		context.Next()
	}
}

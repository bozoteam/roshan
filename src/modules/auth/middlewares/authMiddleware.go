package middlewares

import (
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var jwtKey []byte

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	jwtKey = []byte(os.Getenv("JWT_SECRET"))

	if jwtKey == nil {
		log.Fatal("JWT secrets not set in .env file")
	}
}
func AuthMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")

		if tokenString == "" {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			context.Abort()
			return
		}

		claims := &jwt.StandardClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			context.Abort()
			return
		}

		context.Set("username", claims.Subject)
		context.Next()
	}
}

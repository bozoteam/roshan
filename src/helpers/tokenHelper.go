package helpers

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateToken generates a JWT token
func GenerateToken(subject string, secretKey []byte, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)
	claims := &jwt.StandardClaims{
		Subject:   subject,
		ExpiresAt: expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

package helpers

// GenerateToken generates a JWT token
// func GenerateToken(subject string, secretKey []byte, duration time.Duration) (string, error) {
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
// 		Subject:   subject,
// 		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
// 	})
// 	return token.SignedString(secretKey)
// }

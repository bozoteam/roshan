package helpers

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is not set", key)
	}
	return value
}

func GetEnvAsInt(key string) int64 {
	valueStr := GetEnv(key)
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid integer value for %s: %v", key, err)
	}
	return value
}

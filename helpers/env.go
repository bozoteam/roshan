package helpers

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var development = "true"
var IsDevelopment bool = development == "true"

func LoadDotEnv() {
	file := ".env"
	if _, err := os.Stat(file); err != nil {
		file = ".env.dev"
	}
	if err := godotenv.Load(file); err != nil {
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

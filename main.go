package main

import (
	"log"
	"os"

	"github.com/bozoteam/roshan/src/routes"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	port := os.Getenv("API_PORT")

	router := routes.SetupRouter()
	router.Run(port)
}

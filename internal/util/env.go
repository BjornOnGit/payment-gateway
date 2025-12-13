package util

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	env := os.Getenv("ENV")
	if env == "production" {
		return
	}

	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("No .env.local file found: %v", err)
	}
}
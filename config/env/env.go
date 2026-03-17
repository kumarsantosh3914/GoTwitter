package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func load() {
	err := godotenv.Load()
	if err != nil {
		// Log the error if the .env file not found or cannot be loaded
		log.Println("[ERROR] Something failed")
	}
}

func GetString(key string, fallback string) string {
	load()

	value, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	return value
}

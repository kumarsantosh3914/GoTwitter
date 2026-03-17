package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Load() {
	err := godotenv.Load()
	if err != nil {
		// Log the error if the .env file not found or cannot be loaded
		log.Println("[ERROR] Something failed")
	}
}

func GetString(key string, fallback string) string {
	Load()

	value, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	return value
}

func getInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	intValue, err := strconv.Atoi(value)

	if err != nil {
		log.Printf("[ERROR] Error converting %q to int: %v", key, err)
		return fallback
	}

	return intValue
}

func getBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	boolValue, err := strconv.ParseBool(value)

	if err != nil {
		log.Printf("[ERROR] Error converting %q to bool: %v", key, err)
		return fallback
	}

	return boolValue
}

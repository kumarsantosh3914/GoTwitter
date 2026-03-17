package env

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var loadOnce sync.Once

func Load() {
	loadOnce.Do(func() {
		// Overload so local `.env` always wins over existing env vars.
		if err := godotenv.Overload(".env"); err != nil {
			log.Printf("[WARN] .env not loaded: %v", err)
			return
		}
	})
}

func GetString(key string, fallback string) string {
	Load()

	value, ok := os.LookupEnv(key)

	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

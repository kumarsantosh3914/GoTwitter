package env

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var loadOnce sync.Once

func Load() {
	loadOnce.Do(func() {
		envPath, ok := findEnvFile()
		if !ok {
			log.Printf("[WARN] .env not loaded: %v", os.ErrNotExist)
			return
		}

		if err := loadEnvFile(envPath); err != nil {
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

func findEnvFile() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
// Values are loaded from environment variables with sensible defaults.
// Additional fields (Database, Redis, JWT) will be added in later phases.
type Config struct {
	ServerPort  string
	Environment string
	LogLevel    string
	DatabaseURL string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
	}
}

// getEnv retrieves an environment variable or returns a default value.
// This is a common Go pattern — the standard library doesn't provide
// a "get with default" function for env vars.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

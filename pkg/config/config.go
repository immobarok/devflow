package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
// Values are loaded from environment variables with sensible defaults.
type Config struct {
	ServerPort         string
	Environment        string
	LogLevel           string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	JWTRefreshSecret   string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		LogLevel:           getEnv("LOG_LEVEL", "debug"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		RedisURL:           getEnv("REDIS_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", "devflow-default-jwt-secret-key-min-32-bytes"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", "devflow-default-refresh-secret-key-min-32-bytes"),
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
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

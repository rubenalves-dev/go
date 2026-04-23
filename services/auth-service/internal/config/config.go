package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName       string
	LogLevel          string
	HTTPAddr          string
	DatabaseURL       string
	JWTIssuer         string
	JWTPrivateKeyPath string
	JWTExpiryMinutes  int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ServiceName:       getEnv("SERVICE_NAME", "auth-service"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		HTTPAddr:          getEnv("HTTP_ADDR", ":8081"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTIssuer:         getEnv("JWT_ISSUER", "auth-service"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "./keys/dev-private.pem"),
		JWTExpiryMinutes:  getEnvInt("JWT_EXPIRY_MINUTES", 60),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

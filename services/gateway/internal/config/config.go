package config

import "github.com/joho/godotenv"
import "os"

type Config struct {
	ServiceName    string
	LogLevel       string
	HTTPAddr       string
	AuthServiceURL string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		ServiceName:    getEnv("SERVICE_NAME", "gateway"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

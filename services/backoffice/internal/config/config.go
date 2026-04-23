package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName      string
	LogLevel         string
	HTTPAddr         string
	AdminRoutePrefix string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ServiceName:      getEnv("SERVICE_NAME", "backoffice"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		HTTPAddr:         getEnv("HTTP_ADDR", ":8082"),
		AdminRoutePrefix: getEnv("ADMIN_ROUTE_PREFIX", "/api/v1/admin"),
	}

	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return Config{}, errors.New("HTTP_ADDR cannot be empty")
	}
	if !strings.HasPrefix(cfg.AdminRoutePrefix, "/") {
		return Config{}, errors.New("ADMIN_ROUTE_PREFIX must start with '/'")
	}

	cfg.AdminRoutePrefix = strings.TrimRight(cfg.AdminRoutePrefix, "/")
	if cfg.AdminRoutePrefix == "" {
		return Config{}, errors.New("ADMIN_ROUTE_PREFIX cannot be root")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

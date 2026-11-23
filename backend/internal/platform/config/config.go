package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv            string
	HTTPAddr          string
	LogLevel          string
	DatabaseURL       string
	FrontendDir       string
	AuthTokenTTLHours int
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:            envOrDefault("APP_ENV", "development"),
		HTTPAddr:          envOrDefault("HTTP_ADDR", ":8080"),
		LogLevel:          envOrDefault("LOG_LEVEL", "info"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		FrontendDir:       envOrDefault("FRONTEND_DIR", "../frontend"),
		AuthTokenTTLHours: 72,
	}

	if v := os.Getenv("AUTH_TOKEN_TTL_HOURS"); v != "" {
		hours, err := strconv.Atoi(v)
		if err != nil || hours <= 0 {
			return Config{}, fmt.Errorf("AUTH_TOKEN_TTL_HOURS must be a positive integer")
		}
		cfg.AuthTokenTTLHours = hours
	}

	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

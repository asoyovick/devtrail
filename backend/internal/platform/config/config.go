package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	LogLevel    string
	DatabaseURL string
	FrontendDir string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:      envOrDefault("APP_ENV", "development"),
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		LogLevel:    envOrDefault("LOG_LEVEL", "info"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		FrontendDir: envOrDefault("FRONTEND_DIR", "../frontend"),
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

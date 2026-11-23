package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv             string
	HTTPAddr           string
	LogLevel           string
	DatabaseURL        string
	FrontendDir        string
	AuthTokenTTLHours  int
	CORSAllowedOrigins []string
	MaxBodyBytes       int64
	RateLimitPerMinute int
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:             envOrDefault("APP_ENV", "development"),
		HTTPAddr:           envOrDefault("HTTP_ADDR", ":8080"),
		LogLevel:           envOrDefault("LOG_LEVEL", "info"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		FrontendDir:        envOrDefault("FRONTEND_DIR", "../frontend"),
		AuthTokenTTLHours:  72,
		CORSAllowedOrigins: splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		MaxBodyBytes:       1 << 20,
		RateLimitPerMinute: 120,
	}

	if v := os.Getenv("AUTH_TOKEN_TTL_HOURS"); v != "" {
		hours, err := strconv.Atoi(v)
		if err != nil || hours <= 0 {
			return Config{}, fmt.Errorf("AUTH_TOKEN_TTL_HOURS must be a positive integer")
		}
		cfg.AuthTokenTTLHours = hours
	}
	if v := os.Getenv("MAX_BODY_BYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("MAX_BODY_BYTES must be a positive integer")
		}
		cfg.MaxBodyBytes = n
	}
	if v := os.Getenv("RATE_LIMIT_PER_MINUTE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("RATE_LIMIT_PER_MINUTE must be a positive integer")
		}
		cfg.RateLimitPerMinute = n
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

func splitCSV(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("FRONTEND_DIR", "")
	t.Setenv("AUTH_TOKEN_TTL_HOURS", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("MAX_BODY_BYTES", "")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q", cfg.AppEnv)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q", cfg.LogLevel)
	}
	if cfg.FrontendDir != "../frontend" {
		t.Fatalf("FrontendDir = %q", cfg.FrontendDir)
	}
	if cfg.AuthTokenTTLHours != 72 {
		t.Fatalf("AuthTokenTTLHours = %d", cfg.AuthTokenTTLHours)
	}
	if cfg.MaxBodyBytes != 1<<20 {
		t.Fatalf("MaxBodyBytes = %d", cfg.MaxBodyBytes)
	}
	if cfg.RateLimitPerMinute != 120 {
		t.Fatalf("RateLimitPerMinute = %d", cfg.RateLimitPerMinute)
	}
	if len(cfg.CORSAllowedOrigins) != 0 {
		t.Fatalf("CORSAllowedOrigins = %+v", cfg.CORSAllowedOrigins)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("FRONTEND_DIR", "./web")
	t.Setenv("AUTH_TOKEN_TTL_HOURS", "24")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://example.com")
	t.Setenv("MAX_BODY_BYTES", "2048")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "10")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppEnv != "test" || cfg.HTTPAddr != ":9090" || cfg.LogLevel != "debug" || cfg.DatabaseURL != "postgres://example" || cfg.FrontendDir != "./web" || cfg.AuthTokenTTLHours != 24 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.MaxBodyBytes != 2048 || cfg.RateLimitPerMinute != 10 {
		t.Fatalf("unexpected limits: %+v", cfg)
	}
	if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "http://localhost:3000" {
		t.Fatalf("cors=%+v", cfg.CORSAllowedOrigins)
	}
}

func TestLoadInvalidAuthTokenTTL(t *testing.T) {
	t.Setenv("AUTH_TOKEN_TTL_HOURS", "nope")
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}

package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("FRONTEND_DIR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want development", cfg.AppEnv)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.FrontendDir != "../frontend" {
		t.Fatalf("FrontendDir = %q, want ../frontend", cfg.FrontendDir)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("FRONTEND_DIR", "./web")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "test" || cfg.HTTPAddr != ":9090" || cfg.LogLevel != "debug" || cfg.DatabaseURL != "postgres://example" || cfg.FrontendDir != "./web" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

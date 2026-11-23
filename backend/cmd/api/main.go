package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devtrail/backend/internal/activity"
	"devtrail/backend/internal/platform/config"
	"devtrail/backend/internal/platform/database"
	"devtrail/backend/internal/platform/httpserver"
	"devtrail/backend/internal/platform/logger"
	"devtrail/backend/internal/project"
	"devtrail/backend/internal/session"
	"devtrail/backend/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString("config error: " + err.Error() + "\n")
		os.Exit(1)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		os.Stderr.WriteString("logger error: " + err.Error() + "\n")
		os.Exit(1)
	}
	slog.SetDefault(log)

	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		pool, err = database.Connect(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Error("database connection failed", "error", err)
			os.Exit(1)
		}
		defer pool.Close()
		log.Info("database connected")
	} else {
		log.Warn("DATABASE_URL not set; starting without database")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpserver.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", readyzHandler(pool))
	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpserver.JSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "v1"})
	})

	if pool != nil {
		userStore := user.NewPGStore(pool)
		userService := user.NewService(userStore, time.Duration(cfg.AuthTokenTTLHours)*time.Hour)
		userHandler := user.NewHandler(userService)
		userHandler.RegisterRoutes(mux)
		log.Info("auth routes registered")

		projectStore := project.NewPGStore(pool)
		projectService := project.NewService(projectStore)
		projectHandler := project.NewHandler(projectService, userHandler.RequireAuth)
		projectHandler.RegisterRoutes(mux)
		log.Info("project routes registered")

		sessionStore := session.NewPGStore(pool)
		sessionService := session.NewService(sessionStore)
		sessionHandler := session.NewHandler(sessionService, userHandler.RequireAuth)
		sessionHandler.RegisterRoutes(mux)
		log.Info("session routes registered")

		activityStore := activity.NewPGStore(pool)
		activityService := activity.NewService(activityStore)
		activityHandler := activity.NewHandler(activityService, userHandler.RequireAuth)
		activityHandler.RegisterRoutes(mux)
		log.Info("activity routes registered")
	} else {
		log.Warn("auth routes disabled; DATABASE_URL not configured")
	}

	handler := httpserver.Recover(log, httpserver.RequestID(httpserver.Logging(log, mux)))

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("devtrail api listening", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
	}
}

func readyzHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			httpserver.Error(w, http.StatusServiceUnavailable, "database_not_configured", "database is not configured", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpserver.Error(w, http.StatusServiceUnavailable, "database_unavailable", "database ping failed", nil)
			return
		}

		httpserver.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

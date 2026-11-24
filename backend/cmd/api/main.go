package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"devtrail/backend/internal/activity"
	"devtrail/backend/internal/decision"
	"devtrail/backend/internal/gitactivity"
	"devtrail/backend/internal/insight"
	"devtrail/backend/internal/investigation"
	"devtrail/backend/internal/note"
	"devtrail/backend/internal/platform/config"
	"devtrail/backend/internal/platform/database"
	"devtrail/backend/internal/platform/httpserver"
	"devtrail/backend/internal/platform/logger"
	"devtrail/backend/internal/project"
	"devtrail/backend/internal/report"
	"devtrail/backend/internal/session"
	"devtrail/backend/internal/streaming"
	"devtrail/backend/internal/testrun"
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

		streamHub := streaming.NewHub()
		streamHandler := streaming.NewHandler(streamHub, userHandler.RequireAuth)
		streamHandler.RegisterRoutes(mux)
		log.Info("streaming routes registered")

		activityStore := activity.NewPGStore(pool)
		activityService := activity.NewService(activityStore)
		activityService.SetPublisher(streamPublisher{hub: streamHub})
		activityHandler := activity.NewHandler(activityService, userHandler.RequireAuth)
		activityHandler.RegisterRoutes(mux)
		log.Info("activity routes registered")

		investigationStore := investigation.NewPGStore(pool)
		investigationService := investigation.NewService(investigationStore)
		investigationHandler := investigation.NewHandler(investigationService, userHandler.RequireAuth)
		investigationHandler.RegisterRoutes(mux)
		log.Info("investigation routes registered")

		decisionStore := decision.NewPGStore(pool)
		decisionService := decision.NewService(decisionStore)
		decisionHandler := decision.NewHandler(decisionService, userHandler.RequireAuth)
		decisionHandler.RegisterRoutes(mux)
		log.Info("decision routes registered")

		noteStore := note.NewPGStore(pool)
		noteService := note.NewService(noteStore)
		noteHandler := note.NewHandler(noteService, userHandler.RequireAuth)
		noteHandler.RegisterRoutes(mux)
		log.Info("note routes registered")

		testRunStore := testrun.NewPGStore(pool)
		testRunService := testrun.NewService(testRunStore)
		testRunHandler := testrun.NewHandler(testRunService, userHandler.RequireAuth)
		testRunHandler.RegisterRoutes(mux)
		log.Info("test run routes registered")

		gitStore := gitactivity.NewPGStore(pool)
		gitService := gitactivity.NewService(gitStore)
		gitHandler := gitactivity.NewHandler(gitService, userHandler.RequireAuth)
		gitHandler.RegisterRoutes(mux)
		log.Info("git activity routes registered")

		reportService := report.NewService(pool)
		reportHandler := report.NewHandler(reportService, userHandler.RequireAuth)
		reportHandler.RegisterRoutes(mux)
		log.Info("report routes registered")

		insightService := insight.NewService(pool)
		insightHandler := insight.NewHandler(insightService, userHandler.RequireAuth)
		insightHandler.RegisterRoutes(mux)
		log.Info("insight routes registered")
	} else {
		log.Warn("auth routes disabled; DATABASE_URL not configured")
	}

	if cfg.FrontendDir != "" {
		if info, err := os.Stat(cfg.FrontendDir); err == nil && info.IsDir() {
			fs := http.FileServer(http.Dir(cfg.FrontendDir))
			mux.HandleFunc("GET /", frontendHandler(cfg.FrontendDir, fs))
			log.Info("frontend static files enabled", "dir", cfg.FrontendDir)
		} else {
			log.Warn("frontend directory not found; static serving disabled", "dir", cfg.FrontendDir)
		}
	}

	rateLimiter := httpserver.NewRateLimiter(cfg.RateLimitPerMinute, time.Minute)
	handler := httpserver.BodyLimit(cfg.MaxBodyBytes,
		httpserver.SecurityHeaders(
			httpserver.CORS(cfg.CORSAllowedOrigins,
				rateLimiter.Middleware(
					httpserver.Recover(log, httpserver.RequestID(httpserver.Logging(log, mux))),
				),
			),
		),
	)

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

type streamPublisher struct{ hub *streaming.Hub }

func (p streamPublisher) Publish(userID string, ev activity.PublishedEvent) {
	p.hub.Publish(userID, streaming.Event{ID: ev.ID, Type: ev.Type, ProjectID: ev.ProjectID, SessionID: ev.SessionID, Source: ev.Source, OccurredAt: ev.OccurredAt, Payload: ev.Payload})
}

func frontendHandler(dir string, fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			http.NotFound(w, r)
			return
		}
		clean := filepath.Clean(r.URL.Path)
		if clean == "/" || clean == "." {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		full := filepath.Join(dir, filepath.FromSlash(clean))
		info, err := os.Stat(full)
		if err == nil {
			if info.IsDir() {
				index := filepath.Join(full, "index.html")
				if _, err := os.Stat(index); err == nil {
					http.ServeFile(w, r, index)
					return
				}
				http.NotFound(w, r)
				return
			}
			fs.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}
}

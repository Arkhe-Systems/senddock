package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/arkhe-systems/senddock/internal/cache"
	"github.com/arkhe-systems/senddock/internal/config"
	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/handler"
	"github.com/arkhe-systems/senddock/internal/middleware"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"

	_ "github.com/lib/pq"
)

type App struct {
	cfg config.Config

	conn    *sql.DB
	queries *db.Queries
	cache   *cache.Redis

	mux *http.ServeMux

	authMiddleware   func(http.Handler) http.Handler
	apiKeyMiddleware func(http.Handler) http.Handler
	eitherAuth       func(http.Handler) http.Handler
	rateLimiter      *middleware.RateLimiter

	worker *service.CampaignWorker

	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if cfg.DatabaseUrl == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	conn, err := sql.Open("postgres", cfg.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	log.Println("Connected to PostgreSQL")

	queries := db.New(conn)
	redisCache := cache.NewRedis(cfg.RedisUrl)

	a := &App{
		cfg:     cfg,
		conn:    conn,
		queries: queries,
		cache:   redisCache,
		mux:     http.NewServeMux(),
	}

	a.authMiddleware = middleware.Auth([]byte(cfg.JWTSecret))
	a.apiKeyMiddleware = middleware.APIKey(queries)
	a.eitherAuth = middleware.EitherAuth(a.authMiddleware, a.apiKeyMiddleware)
	a.rateLimiter = middleware.NewRateLimiter(redisCache, 100, time.Minute)

	emailService := service.NewEmailService(queries, cfg.PublicURL, cfg.JWTSecret, redisCache)
	a.worker = service.NewCampaignWorker(queries, emailService)

	a.registerCoreRoutes(emailService)
	a.serveFrontend()

	return a, nil
}

func (a *App) Mux() *http.ServeMux { return a.mux }

func (a *App) DB() *sql.DB { return a.conn }

func (a *App) WithAuth(h http.Handler) http.Handler {
	return a.authMiddleware(h)
}

func (a *App) WithAPIAuth(h http.Handler) http.Handler {
	return a.eitherAuth(h)
}

func (a *App) Run(ctx context.Context) error {
	a.worker.Start()

	wrapped := middleware.Security(
		middleware.LimitBody(
			a.rateLimiter.Middleware(
				middleware.CORS(a.cfg.FrontendURL)(a.mux),
			),
		),
	)

	a.server = &http.Server{
		Addr:         ":" + a.cfg.Port,
		Handler:      wrapped,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Println("Server running:" + a.cfg.Port)
		err := a.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received, draining...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return a.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (a *App) Close() error {
	if a.cache != nil {
		a.cache.Close()
	}
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

func (a *App) registerCoreRoutes(emailService *service.EmailService) {
	cfg := a.cfg
	queries := a.queries

	authService := service.NewAuthService(queries, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	projectService := service.NewProjectService(queries, cfg.JWTSecret)
	projectHandler := handler.NewProjectHandler(projectService)

	subscriberService := service.NewSubscriberService(queries)
	subscriberHandler := handler.NewSubscriberHandler(subscriberService, projectService)

	templateService := service.NewTemplateService(queries)
	templateHandler := handler.NewTemplateHandler(templateService, projectService)

	apiKeyService := service.NewAPIKeyService(queries)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService, projectService)

	emailHandler := handler.NewEmailHandler(emailService, projectService, a.cache)

	campaignService := service.NewCampaignService(queries)
	campaignHandler := handler.NewCampaignHandler(campaignService, projectService, cfg.PublicURL)

	trackingHandler := handler.NewTrackingHandler(queries)

	releaseService := service.NewReleaseService(a.cache)
	releaseHandler := handler.NewReleaseHandler(releaseService)

	waitlistHandler := handler.NewWaitlistHandler(subscriberService, emailService)
	setupHandler := handler.NewSetupHandler(queries, authService, cfg)

	mux := a.mux
	authMW := a.authMiddleware
	eitherAuth := a.eitherAuth

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/v1/setup/status", setupHandler.Status)
	mux.HandleFunc("POST /api/v1/setup", setupHandler.Setup)

	mux.Handle("GET /api/v1/me", authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(string)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
	})))

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	if cfg.IsSelfHosted() {
		log.Println("Mode: self-hosted (registration disabled)")
	} else {
		log.Println("Mode: cloud (registration enabled)")
		mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	}

	mux.Handle("POST /api/v1/projects", authMW(http.HandlerFunc(projectHandler.Create)))
	mux.Handle("GET /api/v1/projects", authMW(http.HandlerFunc(projectHandler.List)))
	mux.Handle("GET /api/v1/projects/{id}", authMW(http.HandlerFunc(projectHandler.Get)))
	mux.Handle("PUT /api/v1/projects/{id}", authMW(http.HandlerFunc(projectHandler.Update)))
	mux.Handle("DELETE /api/v1/projects/{id}", authMW(http.HandlerFunc(projectHandler.Delete)))
	mux.Handle("PUT /api/v1/projects/{id}/smtp", authMW(http.HandlerFunc(projectHandler.UpdateSMTP)))

	mux.Handle("POST /api/v1/projects/{id}/subscribers", authMW(http.HandlerFunc(subscriberHandler.Create)))
	mux.Handle("GET /api/v1/projects/{id}/subscribers", authMW(http.HandlerFunc(subscriberHandler.List)))
	mux.Handle("POST /api/v1/projects/{id}/subscribers/bulk", authMW(http.HandlerFunc(subscriberHandler.BulkAction)))
	mux.Handle("PATCH /api/v1/projects/{id}/subscribers/{subscriberId}", authMW(http.HandlerFunc(subscriberHandler.UpdateStatus)))
	mux.Handle("DELETE /api/v1/projects/{id}/subscribers/{subscriberId}", authMW(http.HandlerFunc(subscriberHandler.Delete)))
	mux.Handle("POST /api/v1/projects/{id}/subscribers/import", eitherAuth(http.HandlerFunc(subscriberHandler.Import)))

	mux.Handle("POST /api/v1/projects/{id}/keys", authMW(http.HandlerFunc(apiKeyHandler.Create)))
	mux.Handle("GET /api/v1/projects/{id}/keys", authMW(http.HandlerFunc(apiKeyHandler.List)))
	mux.Handle("DELETE /api/v1/projects/{id}/keys/{keyId}", authMW(http.HandlerFunc(apiKeyHandler.Delete)))

	mux.Handle("POST /api/v1/projects/{id}/templates", authMW(http.HandlerFunc(templateHandler.Create)))
	mux.Handle("GET /api/v1/projects/{id}/templates", authMW(http.HandlerFunc(templateHandler.List)))
	mux.Handle("GET /api/v1/projects/{id}/templates/{templateId}", authMW(http.HandlerFunc(templateHandler.Get)))
	mux.Handle("PUT /api/v1/projects/{id}/templates/{templateId}", authMW(http.HandlerFunc(templateHandler.Update)))
	mux.Handle("DELETE /api/v1/projects/{id}/templates/{templateId}", authMW(http.HandlerFunc(templateHandler.Delete)))

	mux.Handle("POST /api/v1/projects/{id}/campaigns", authMW(http.HandlerFunc(campaignHandler.Create)))
	mux.Handle("GET /api/v1/projects/{id}/campaigns", authMW(http.HandlerFunc(campaignHandler.List)))
	mux.Handle("DELETE /api/v1/projects/{id}/campaigns/{campaignId}", authMW(http.HandlerFunc(campaignHandler.Delete)))
	mux.Handle("PATCH /api/v1/projects/{id}/campaigns/{campaignId}", authMW(http.HandlerFunc(campaignHandler.Update)))

	mux.Handle("POST /api/v1/projects/{id}/smtp/test", authMW(http.HandlerFunc(emailHandler.TestSMTP)))
	mux.Handle("POST /api/v1/projects/{id}/send", eitherAuth(http.HandlerFunc(emailHandler.Send)))
	mux.Handle("POST /api/v1/projects/{id}/broadcast", eitherAuth(http.HandlerFunc(emailHandler.Broadcast)))
	mux.Handle("POST /api/v1/projects/{id}/send/batch", eitherAuth(http.HandlerFunc(emailHandler.BatchSend)))
	mux.Handle("GET /api/v1/projects/{id}/logs", authMW(http.HandlerFunc(emailHandler.Logs)))
	mux.Handle("GET /api/v1/projects/{id}/stats", eitherAuth(http.HandlerFunc(emailHandler.Stats)))

	mux.HandleFunc("GET /unsubscribe/{id}/{subscriberId}", emailHandler.UnsubscribePage)
	mux.HandleFunc("POST /unsubscribe/{id}/{subscriberId}", emailHandler.Unsubscribe)

	mux.HandleFunc("GET /t/{logId}", trackingHandler.Open)
	mux.HandleFunc("POST /api/v1/projects/{id}/waitlist", waitlistHandler.Join)
	mux.HandleFunc("OPTIONS /api/v1/projects/{id}/waitlist", waitlistHandler.Join)

	mux.Handle("GET /api/v1/version", authMW(http.HandlerFunc(releaseHandler.Get)))

	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
}

func (a *App) serveFrontend() {
	distPath := os.Getenv("FRONTEND_DIST_PATH")
	if distPath == "" {
		distPath = "../frontend/dist"
	}

	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		log.Println("Frontend dist/ not found, skipping static file serving")
		return
	}

	frontendFS := os.DirFS(distPath)
	fileServer := http.FileServerFS(frontendFS)

	a.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/unsubscribe/") || strings.HasPrefix(r.URL.Path, "/t/") || r.URL.Path == "/health" {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(frontendFS, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		indexFile, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexFile)
	})

	log.Println("Serving frontend from " + distPath)
}

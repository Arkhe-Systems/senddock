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
	"sync/atomic"
	"time"

	"github.com/arkhe-systems/senddock/internal/cache"
	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/handler"
	"github.com/arkhe-systems/senddock/internal/middleware"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/internal/webhooks"
	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/arkhe-systems/senddock/pkg/config"
	"github.com/arkhe-systems/senddock/pkg/license"

	"github.com/google/uuid"
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

	worker          *service.CampaignWorker
	broadcastWorker *service.BroadcastWorker
	webhooks     *webhooks.Service
	suppressions *service.SuppressionService
	audit        *service.AuditService
	bouncePoller *service.BounceIMAPPoller
	workspaces   *service.WorkspaceService
	projects     *service.ProjectService
	watchtower   *service.WatchtowerClient

	server *http.Server

	dbHealthLastSuccess atomic.Int64
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
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)
	conn.SetConnMaxIdleTime(2 * time.Minute)
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
	a.rateLimiter = middleware.NewRateLimiter(redisCache, cfg.RateLimitPerMinute, time.Minute)

	a.webhooks = webhooks.NewService(queries)
	suppressionService := service.NewSuppressionService(queries)
	emailService := service.NewEmailService(queries, cfg.PublicURL, cfg.JWTSecret, redisCache, a.webhooks, suppressionService)
	a.worker = service.NewCampaignWorker(queries, emailService)
	a.broadcastWorker = service.NewBroadcastWorker(queries, emailService)
	a.suppressions = suppressionService
	a.audit = service.NewAuditService(queries)
	a.bouncePoller = service.NewBounceIMAPPoller(queries, suppressionService, cfg.JWTSecret)
	a.watchtower = service.NewWatchtowerClient(cfg.WatchtowerURL, cfg.WatchtowerToken)
	if a.watchtower != nil {
		probeCtx, cancelProbe := context.WithTimeout(context.Background(), 5*time.Second)
		if err := a.watchtower.Ping(probeCtx); err != nil {
			log.Printf("watchtower: probe failed (will retry on request): %v", err)
		} else {
			log.Printf("watchtower: reachable at %s", cfg.WatchtowerURL)
		}
		cancelProbe()
	}

	recoveryCtx, cancelRecovery := context.WithTimeout(context.Background(), 5*time.Second)
	if err := emailService.RecoverInProgressBroadcasts(recoveryCtx); err != nil {
		log.Printf("broadcast recovery: %v", err)
	}
	cancelRecovery()

	a.registerCoreRoutes(emailService)
	a.serveFrontend()

	return a, nil
}

func (a *App) Mux() *http.ServeMux { return a.mux }

func (a *App) DB() *sql.DB { return a.conn }

func (a *App) Webhooks() *webhooks.Service { return a.webhooks }

func (a *App) Audit() *service.AuditService { return a.audit }

func (a *App) Projects() *service.ProjectService { return a.projects }

func (a *App) SetLicenseGate(gate license.Gate) {
	if a.workspaces != nil {
		a.workspaces.SetLicenseGate(gate)
	}
}

func (a *App) WithAuth(h http.Handler) http.Handler {
	return a.authMiddleware(h)
}

func (a *App) WithAPIAuth(h http.Handler) http.Handler {
	return a.eitherAuth(h)
}

func (a *App) Run(ctx context.Context) error {
	a.startDBHealthMonitor(ctx)
	a.worker.Start()
	a.broadcastWorker.Start(ctx)
	a.webhooks.Start(ctx)
	a.bouncePoller.Start(ctx)

	wrapped := middleware.Security(
		middleware.LimitBody(
			a.rateLimiter.Middleware(
				middleware.CORS(a.cfg.FrontendURL)(a.mux),
			),
		),
	)

	rootMux := http.NewServeMux()
	rootMux.HandleFunc("GET /health", a.healthHandler)
	rootMux.Handle("/", wrapped)

	a.server = &http.Server{
		Addr:         "0.0.0.0:" + a.cfg.Port,
		Handler:      rootMux,
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

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	last := a.dbHealthLastSuccess.Load()
	ageSec := time.Now().Unix() - last
	if last == 0 || ageSec > 60 {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{"status": "db_unreachable", "last_db_check_age_s": ageSec})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"status": "ok", "last_db_check_age_s": ageSec})
}

func (a *App) startDBHealthMonitor(ctx context.Context) {
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := a.conn.PingContext(pingCtx); err == nil {
		a.dbHealthLastSuccess.Store(time.Now().Unix())
	} else {
		log.Printf("startup db ping failed: %v", err)
	}
	cancel()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pctx, pcancel := context.WithTimeout(context.Background(), 5*time.Second)
				err := a.conn.PingContext(pctx)
				pcancel()
				if err != nil {
					log.Printf("background db ping failed: %v", err)
					continue
				}
				a.dbHealthLastSuccess.Store(time.Now().Unix())
			}
		}
	}()
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
	a.projects = projectService
	projectHandler := handler.NewProjectHandler(projectService)

	workspaceService := service.NewWorkspaceService(queries)
	a.workspaces = workspaceService
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService, projectService)

	emailValidator := service.NewEmailValidator()
	subscriberService := service.NewSubscriberService(queries, a.webhooks, emailValidator, a.suppressions)
	suppressionHandler := handler.NewSuppressionHandler(a.suppressions, projectService)
	subscriberHandler := handler.NewSubscriberHandler(subscriberService, projectService)

	templateService := service.NewTemplateService(queries)
	templateHandler := handler.NewTemplateHandler(templateService, projectService)

	apiKeyService := service.NewAPIKeyService(queries)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService, projectService)

	emailHandler := handler.NewEmailHandler(emailService, projectService, a.cache)

	campaignService := service.NewCampaignService(queries)
	campaignHandler := handler.NewCampaignHandler(campaignService, projectService, a.worker, cfg.PublicURL)

	trackingHandler := handler.NewTrackingHandler(queries, emailService, a.webhooks)

	projectHandler.Audit = a.audit
	subscriberHandler.Audit = a.audit
	apiKeyHandler.Audit = a.audit
	emailHandler.Audit = a.audit
	suppressionHandler.Audit = a.audit
	workspaceHandler.Audit = a.audit

	releaseService := service.NewReleaseService(a.cache, cfg.DeploymentMode)
	releaseHandler := handler.NewReleaseHandler(releaseService)

	waitlistHandler := handler.NewWaitlistHandler(subscriberService, emailService)
	setupHandler := handler.NewSetupHandler(queries, authService, cfg)

	mux := a.mux
	authMW := a.authMiddleware
	eitherAuth := a.eitherAuth

	mux.HandleFunc("GET /api/v1/setup/status", setupHandler.Status)
	mux.HandleFunc("POST /api/v1/setup", setupHandler.Setup)

	mux.Handle("GET /api/v1/me", authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(string)
		w.Header().Set("Content-Type", "application/json")

		uid, err := uuid.Parse(userID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid user id"})
			return
		}
		user, err := a.queries.GetUserById(r.Context(), uid)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"user_id":    userID,
			"email":      user.Email,
			"name":       user.Name,
			"plan":       user.Plan,
			"created_at": user.CreatedAt,
		})
	})))

	mux.Handle("POST /api/v1/me/password", authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(string)
		w.Header().Set("Content-Type", "application/json")

		uid, err := uuid.Parse(userID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid user id"})
			return
		}

		var req struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}
		if req.CurrentPassword == "" || req.NewPassword == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "current_password and new_password are required"})
			return
		}

		if err := authService.ChangePassword(r.Context(), uid, req.CurrentPassword, req.NewPassword); err != nil {
			status := http.StatusBadRequest
			if err.Error() == "current password is incorrect" {
				status = http.StatusUnauthorized
			}
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "password updated"})
	})))

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	if cfg.IsSelfHosted() {
		log.Println("Mode: self-hosted (registration disabled)")
	} else {
		log.Println("Mode: cloud (registration enabled)")
		mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	}

	mux.Handle("POST /api/v1/workspaces", authMW(http.HandlerFunc(workspaceHandler.Create)))
	mux.Handle("GET /api/v1/workspaces", authMW(http.HandlerFunc(workspaceHandler.List)))
	mux.Handle("PATCH /api/v1/workspaces/{id}", authMW(http.HandlerFunc(workspaceHandler.Rename)))
	mux.Handle("DELETE /api/v1/workspaces/{id}", authMW(http.HandlerFunc(workspaceHandler.Delete)))
	mux.Handle("GET /api/v1/workspaces/{id}/projects", authMW(http.HandlerFunc(workspaceHandler.ListProjects)))
	mux.Handle("GET /api/v1/workspaces/{id}/members", authMW(http.HandlerFunc(workspaceHandler.ListMembers)))
	mux.Handle("POST /api/v1/workspaces/{id}/members", authMW(http.HandlerFunc(workspaceHandler.AddMember)))
	mux.Handle("POST /api/v1/workspaces/{id}/users", authMW(http.HandlerFunc(workspaceHandler.CreateUser)))
	mux.Handle("PATCH /api/v1/workspaces/{id}/members/{userId}", authMW(http.HandlerFunc(workspaceHandler.UpdateMember)))
	mux.Handle("DELETE /api/v1/workspaces/{id}/members/{userId}", authMW(http.HandlerFunc(workspaceHandler.RemoveMember)))

	mux.Handle("POST /api/v1/projects", authMW(http.HandlerFunc(projectHandler.Create)))
	mux.Handle("GET /api/v1/projects", authMW(http.HandlerFunc(projectHandler.List)))
	mux.Handle("GET /api/v1/projects/{id}", authMW(http.HandlerFunc(projectHandler.Get)))
	mux.Handle("PUT /api/v1/projects/{id}", authMW(http.HandlerFunc(projectHandler.Update)))
	mux.Handle("DELETE /api/v1/projects/{id}", authMW(http.HandlerFunc(projectHandler.Delete)))
	mux.Handle("PUT /api/v1/projects/{id}/smtp", authMW(http.HandlerFunc(projectHandler.UpdateSMTP)))
	mux.Handle("GET /api/v1/projects/{id}/bounce-webhook", authMW(http.HandlerFunc(projectHandler.GetBounceWebhook)))
	mux.Handle("POST /api/v1/projects/{id}/bounce-webhook/rotate", authMW(http.HandlerFunc(projectHandler.RotateBounceToken)))
	mux.Handle("GET /api/v1/projects/{id}/bounce-imap", authMW(http.HandlerFunc(projectHandler.GetBounceIMAP)))
	mux.Handle("PUT /api/v1/projects/{id}/bounce-imap", authMW(http.HandlerFunc(projectHandler.UpdateBounceIMAP)))

	bounceWebhookHandler := handler.NewBounceWebhookHandler(queries, a.suppressions)
	mux.HandleFunc("POST /webhooks/bounces/{projectId}", bounceWebhookHandler.Receive)

	mux.Handle("POST /api/v1/projects/{id}/subscribers", authMW(http.HandlerFunc(subscriberHandler.Create)))
	mux.Handle("GET /api/v1/projects/{id}/subscribers", authMW(http.HandlerFunc(subscriberHandler.List)))
	mux.Handle("POST /api/v1/projects/{id}/subscribers/bulk", authMW(http.HandlerFunc(subscriberHandler.BulkAction)))
	mux.Handle("PATCH /api/v1/projects/{id}/subscribers/{subscriberId}", authMW(http.HandlerFunc(subscriberHandler.UpdateStatus)))
	mux.Handle("DELETE /api/v1/projects/{id}/subscribers/{subscriberId}", authMW(http.HandlerFunc(subscriberHandler.Delete)))
	mux.Handle("POST /api/v1/projects/{id}/subscribers/import", eitherAuth(http.HandlerFunc(subscriberHandler.Import)))

	mux.Handle("GET /api/v1/projects/{id}/suppressions", authMW(http.HandlerFunc(suppressionHandler.List)))
	mux.Handle("POST /api/v1/projects/{id}/suppressions", authMW(http.HandlerFunc(suppressionHandler.Add)))
	mux.Handle("DELETE /api/v1/projects/{id}/suppressions/{suppressionId}", authMW(http.HandlerFunc(suppressionHandler.Delete)))

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
	mux.Handle("GET /api/v1/projects/{id}/logs/export.csv", authMW(http.HandlerFunc(emailHandler.LogsExport)))
	mux.Handle("GET /api/v1/projects/{id}/broadcasts", authMW(http.HandlerFunc(emailHandler.ListBroadcasts)))
	mux.Handle("GET /api/v1/projects/{id}/stats", eitherAuth(http.HandlerFunc(emailHandler.Stats)))

	mux.HandleFunc("GET /unsubscribe/{id}/{subscriberId}", emailHandler.UnsubscribePage)
	mux.HandleFunc("POST /unsubscribe/{id}/{subscriberId}", emailHandler.Unsubscribe)

	mux.HandleFunc("GET /t/{logId}", trackingHandler.Open)
	mux.HandleFunc("GET /c/{logId}/{payload}", trackingHandler.Click)
	mux.HandleFunc("POST /api/v1/projects/{id}/waitlist", waitlistHandler.Join)
	mux.HandleFunc("OPTIONS /api/v1/projects/{id}/waitlist", waitlistHandler.Join)

	mux.Handle("GET /api/v1/version", authMW(http.HandlerFunc(releaseHandler.Get)))

	mux.Handle("GET /api/v1/update/auto-status", authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(a.watchtower.StatusFresh(r.Context()))
	})))

	mux.Handle("POST /api/v1/update/trigger", authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !a.watchtower.Configured() {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "watchtower not configured"})
			return
		}
		status := a.watchtower.StatusFresh(r.Context())
		if !status.Healthy {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": "watchtower unreachable: " + status.LastError})
			return
		}

		go func() {
			triggerCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			if err := a.watchtower.Trigger(triggerCtx); err != nil {
				log.Printf("watchtower trigger failed: %v", err)
			} else {
				log.Println("watchtower trigger accepted")
			}
		}()

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"message": "update triggered"})
	})))

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
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/unsubscribe/") || strings.HasPrefix(r.URL.Path, "/t/") || strings.HasPrefix(r.URL.Path, "/c/") || strings.HasPrefix(r.URL.Path, "/webhooks/") || r.URL.Path == "/health" {
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

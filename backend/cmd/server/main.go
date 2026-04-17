package main

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/arkhe-systems/senddock/internal/config"
	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/handler"
	"github.com/arkhe-systems/senddock/internal/middleware"
	"github.com/arkhe-systems/senddock/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	conn, err := sql.Open("postgres", cfg.DatabaseUrl)
	if err != nil {
		log.Fatal("Error connecting to database ", err)
	}

	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("Error pinging database: ", err)
	}
	log.Println("Connected to PostgreSQL")
	queries := db.New(conn)

	authService := service.NewAuthService(queries, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	projectService := service.NewProjectService(queries)
	projectHandler := handler.NewProjectHandler(projectService)

	subscriberService := service.NewSubscriberService(queries)
	subscriberHandler := handler.NewSubscriberHandler(subscriberService, projectService)

	templateService := service.NewTemplateService(queries)
	templateHandler := handler.NewTemplateHandler(templateService, projectService)

	apiKeyService := service.NewAPIKeyService(queries)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService, projectService)

	emailService := service.NewEmailService(queries, cfg.FrontendURL)
	emailHandler := handler.NewEmailHandler(emailService, projectService)

	setupHandler := handler.NewSetupHandler(queries, authService, cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	authMiddleware := middleware.Auth([]byte(cfg.JWTSecret))
	apiKeyMiddleware := middleware.APIKey(queries)
	eitherAuth := middleware.EitherAuth(authMiddleware, apiKeyMiddleware)

	mux.HandleFunc("GET /api/v1/setup/status", setupHandler.Status)
	mux.HandleFunc("POST /api/v1/setup", setupHandler.Setup)

	mux.Handle("GET /api/v1/me", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).(string)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"user_id": userID,
		})
	})))

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	if cfg.IsSelfHosted() {
		log.Println("Mode: self-hosted (registration disabled)")
	} else {
		log.Println("Mode: cloud (registration enabled)")
		mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	}

	mux.Handle("POST /api/v1/projects", authMiddleware(http.HandlerFunc(projectHandler.Create)))
	mux.Handle("GET /api/v1/projects", authMiddleware(http.HandlerFunc(projectHandler.List)))
	mux.Handle("GET /api/v1/projects/{id}", authMiddleware(http.HandlerFunc(projectHandler.Get)))
	mux.Handle("PUT /api/v1/projects/{id}", authMiddleware(http.HandlerFunc(projectHandler.Update)))
	mux.Handle("DELETE /api/v1/projects/{id}", authMiddleware(http.HandlerFunc(projectHandler.Delete)))
	mux.Handle("PUT /api/v1/projects/{id}/smtp", authMiddleware(http.HandlerFunc(projectHandler.UpdateSMTP)))

	mux.Handle("POST /api/v1/projects/{id}/subscribers", authMiddleware(http.HandlerFunc(subscriberHandler.Create)))
	mux.Handle("GET /api/v1/projects/{id}/subscribers", authMiddleware(http.HandlerFunc(subscriberHandler.List)))
	mux.Handle("PATCH /api/v1/projects/{id}/subscribers/{subscriberId}", authMiddleware(http.HandlerFunc(subscriberHandler.UpdateStatus)))
	mux.Handle("DELETE /api/v1/projects/{id}/subscribers/{subscriberId}", authMiddleware(http.HandlerFunc(subscriberHandler.Delete)))

	mux.Handle("POST /api/v1/projects/{id}/keys", authMiddleware(http.HandlerFunc(apiKeyHandler.Create)))
	mux.Handle("GET /api/v1/projects/{id}/keys", authMiddleware(http.HandlerFunc(apiKeyHandler.List)))
	mux.Handle("DELETE /api/v1/projects/{id}/keys/{keyId}", authMiddleware(http.HandlerFunc(apiKeyHandler.Delete)))

	mux.Handle("POST /api/v1/projects/{id}/templates", authMiddleware(http.HandlerFunc(templateHandler.Create)))
	mux.Handle("GET /api/v1/projects/{id}/templates", authMiddleware(http.HandlerFunc(templateHandler.List)))
	mux.Handle("GET /api/v1/projects/{id}/templates/{templateId}", authMiddleware(http.HandlerFunc(templateHandler.Get)))
	mux.Handle("PUT /api/v1/projects/{id}/templates/{templateId}", authMiddleware(http.HandlerFunc(templateHandler.Update)))
	mux.Handle("DELETE /api/v1/projects/{id}/templates/{templateId}", authMiddleware(http.HandlerFunc(templateHandler.Delete)))

	mux.Handle("POST /api/v1/projects/{id}/smtp/test", authMiddleware(http.HandlerFunc(emailHandler.TestSMTP)))
	mux.Handle("POST /api/v1/projects/{id}/send", eitherAuth(http.HandlerFunc(emailHandler.Send)))
	mux.Handle("POST /api/v1/projects/{id}/broadcast", eitherAuth(http.HandlerFunc(emailHandler.Broadcast)))
	mux.Handle("GET /api/v1/projects/{id}/logs", authMiddleware(http.HandlerFunc(emailHandler.Logs)))
	mux.Handle("GET /api/v1/projects/{id}/stats", eitherAuth(http.HandlerFunc(emailHandler.Stats)))

	mux.HandleFunc("GET /unsubscribe/{id}/{subscriberId}", emailHandler.Unsubscribe)

	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	serveFrontend(mux)

	log.Println("Server running:" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, middleware.CORS(cfg.FrontendURL)(mux)))
}

func serveFrontend(mux *http.ServeMux) {
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/unsubscribe/") || r.URL.Path == "/health" {
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

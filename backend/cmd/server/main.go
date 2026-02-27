package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

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

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	authMiddleware := middleware.Auth([]byte(cfg.JWTSecret))

	mux.Handle("GET /api/v1/me", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).(string)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"user_id": userID,
		})
	})))

	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	mux.Handle("POST /api/v1/projects", authMiddleware(http.HandlerFunc(projectHandler.Create)))
	mux.Handle("GET /api/v1/projects", authMiddleware(http.HandlerFunc(projectHandler.List)))
	mux.Handle("GET /api/v1/projects/{id}", authMiddleware(http.HandlerFunc(projectHandler.Get)))
	mux.Handle("DELETE /api/v1/projects/{id}", authMiddleware(http.HandlerFunc(projectHandler.Delete)))

	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	log.Println("Server running:" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, middleware.CORS(cfg.FrontendURL)(mux)))

}

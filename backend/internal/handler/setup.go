package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/internal/settings"
	"github.com/arkhe-systems/senddock/pkg/config"
)

type SetupHandler struct {
	queries     *db.Queries
	authService *service.AuthService
	cfg         config.Config
	settings    *settings.Provider
}

func NewSetupHandler(queries *db.Queries, authService *service.AuthService, cfg config.Config, settings *settings.Provider) *SetupHandler {
	return &SetupHandler{
		queries:     queries,
		authService: authService,
		cfg:         cfg,
		settings:    settings,
	}
}

func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	count, _ := h.queries.CountUsers(r.Context())
	current := h.settings.Current()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"setup_required":               count == 0,
		"deployment_mode":              h.cfg.DeploymentModeName(),
		"public_url":                   current.PublicURL,
		"session_idle_timeout_minutes": current.SessionIdleTimeoutMinutes,
	})
}

type setupRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	PublicURL string `json:"public_url"`
}

func (h *SetupHandler) Setup(w http.ResponseWriter, r *http.Request) {
	count, _ := h.queries.CountUsers(r.Context())
	if count > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(errorResponse{Error: "setup already completed"})
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "name, email, and password are required"})
		return
	}

	if publicURL := strings.TrimSpace(req.PublicURL); publicURL != "" {
		current := h.settings.Current()
		current.PublicURL = publicURL
		if _, err := h.settings.Update(r.Context(), current); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	tokens, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	setAuthCookies(w, tokens)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "setup complete"})
}

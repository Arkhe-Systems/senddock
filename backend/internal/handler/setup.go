package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/config"
	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
)

type SetupHandler struct {
	queries     *db.Queries
	authService *service.AuthService
	cfg         config.Config
}

func NewSetupHandler(queries *db.Queries, authService *service.AuthService, cfg config.Config) *SetupHandler {
	return &SetupHandler{
		queries:     queries,
		authService: authService,
		cfg:         cfg,
	}
}

func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	count, _ := h.queries.CountUsers(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"setup_required":  count == 0,
		"deployment_mode": h.cfg.DeploymentMode,
	})
}

type setupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
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

	tokens, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	setAuthCookies(w, tokens)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "setup complete"})
}

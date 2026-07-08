package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/internal/settings"
	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/google/uuid"
)

type InstanceSettingsHandler struct {
	queries  *db.Queries
	provider *settings.Provider
	Audit    *service.AuditService
}

func NewInstanceSettingsHandler(queries *db.Queries, provider *settings.Provider) *InstanceSettingsHandler {
	return &InstanceSettingsHandler{queries: queries, provider: provider}
}

type instanceSettingsResponse struct {
	PublicURL                 string `json:"public_url"`
	SessionIdleTimeoutMinutes int    `json:"session_idle_timeout_minutes"`
}

type updateInstanceSettingsRequest struct {
	PublicURL                 *string `json:"public_url"`
	SessionIdleTimeoutMinutes *int    `json:"session_idle_timeout_minutes"`
}

func (h *InstanceSettingsHandler) requireOwner(r *http.Request) (string, bool) {
	userID, _ := r.Context().Value(auth.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", false
	}
	owned, err := h.queries.CountOwnedWorkspacesByUser(r.Context(), uid)
	if err != nil || owned == 0 {
		return "", false
	}
	return userID, true
}

func (h *InstanceSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireOwner(r); !ok {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "only a workspace owner can view instance settings"})
		return
	}

	current := h.provider.Current()
	writeJSON(w, http.StatusOK, instanceSettingsResponse{
		PublicURL:                 current.PublicURL,
		SessionIdleTimeoutMinutes: current.SessionIdleTimeoutMinutes,
	})
}

func (h *InstanceSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireOwner(r)
	if !ok {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "only a workspace owner can change instance settings"})
		return
	}

	var req updateInstanceSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	next := h.provider.Current()
	if req.PublicURL != nil {
		next.PublicURL = *req.PublicURL
	}
	if req.SessionIdleTimeoutMinutes != nil {
		next.SessionIdleTimeoutMinutes = *req.SessionIdleTimeoutMinutes
	}

	updated, err := h.provider.Update(r.Context(), next)
	if err != nil {
		if errors.Is(err, settings.ErrInvalidIdleTimeout) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not save instance settings"})
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, "", userID, "instance.settings.updated", "instance", "settings", map[string]any{
			"public_url":                   updated.PublicURL,
			"session_idle_timeout_minutes": updated.SessionIdleTimeoutMinutes,
		})
	}

	writeJSON(w, http.StatusOK, instanceSettingsResponse{
		PublicURL:                 updated.PublicURL,
		SessionIdleTimeoutMinutes: updated.SessionIdleTimeoutMinutes,
	})
}

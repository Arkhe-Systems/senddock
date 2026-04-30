package handler

import (
	"errors"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
)

func requireCap(w http.ResponseWriter, r *http.Request, projects *service.ProjectService, cap service.Capability) (string, string, bool) {
	if pid, ok := r.Context().Value(auth.ProjectIDKey).(string); ok {
		return pid, "", true
	}
	userID, _ := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	if userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "missing user")
		return "", "", false
	}
	if err := projects.RequireCapability(r.Context(), projectID, userID, string(cap)); err != nil {
		switch {
		case errors.Is(err, service.ErrCapabilityForbidden):
			writeJSONError(w, http.StatusForbidden, "your role does not allow this action")
		default:
			writeJSONError(w, http.StatusNotFound, "project not found")
		}
		return "", "", false
	}
	return projectID, userID, true
}

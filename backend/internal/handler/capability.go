package handler

import (
	"errors"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
)

// apiKeyCapabilities is the fixed set of capabilities an API key may exercise.
// API keys are project-scoped machine credentials, not user roles — they can
// send and manage subscribers programmatically, but must never be an implicit
// pass for capabilities they were not meant to reach (e.g. member management,
// project settings) if such an endpoint is ever mounted under eitherAuth.
var apiKeyCapabilities = map[service.Capability]bool{
	service.CapSendTransactional: true,
	service.CapBroadcast:         true,
	service.CapSubscribersWrite:  true,
}

func requireCap(w http.ResponseWriter, r *http.Request, projects *service.ProjectService, cap service.Capability) (string, string, bool) {
	if pid, ok := r.Context().Value(auth.ProjectIDKey).(string); ok {
		if !apiKeyCapabilities[cap] {
			writeJSONError(w, http.StatusForbidden, "this action is not permitted with an API key")
			return "", "", false
		}
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

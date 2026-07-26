package handler

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/google/uuid"
)

const bounceWebhookMaxBody = 65536

type BounceWebhookHandler struct {
	queries      *db.Queries
	suppressions *service.SuppressionService
}

func NewBounceWebhookHandler(queries *db.Queries, suppressions *service.SuppressionService) *BounceWebhookHandler {
	return &BounceWebhookHandler{queries: queries, suppressions: suppressions}
}

type genericBouncePayload struct {
	Email  string `json:"email"`
	Reason string `json:"reason"`
	Type   string `json:"type"`
}

type mailgunPayload struct {
	EventData mailgunEventData `json:"event-data"`
}

type mailgunEventData struct {
	Event     string                 `json:"event"`
	Recipient string                 `json:"recipient"`
	Severity  string                 `json:"severity"`
	Reason    string                 `json:"reason"`
	Delivery  map[string]interface{} `json:"delivery-status"`
}

func (h *BounceWebhookHandler) Receive(w http.ResponseWriter, r *http.Request) {
	pid, err := uuid.Parse(r.PathValue("projectId"))
	if err != nil {
		writeBounceError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	tokenStr := r.URL.Query().Get("token")
	tokenUUID, err := uuid.Parse(tokenStr)
	if err != nil {
		writeBounceError(w, http.StatusUnauthorized, "missing or invalid token")
		return
	}

	project, err := h.queries.GetProjectByBounceToken(r.Context(), db.GetProjectByBounceTokenParams{
		ID:          pid,
		BounceToken: tokenUUID,
	})
	if err != nil {
		writeBounceError(w, http.StatusUnauthorized, "invalid project or token")
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, bounceWebhookMaxBody+1))
	if err != nil {
		writeBounceError(w, http.StatusBadRequest, "could not read body")
		return
	}
	if len(bodyBytes) > bounceWebhookMaxBody {
		writeBounceError(w, http.StatusRequestEntityTooLarge, "payload too large")
		return
	}

	email, reason, ok := extractBounce(bodyBytes)
	if !ok {
		writeBounceError(w, http.StatusBadRequest, "could not find email in payload")
		return
	}

	if h.suppressions != nil {
		_, _ = h.suppressions.Add(r.Context(), project.ID, email, service.SuppressionReasonBounce, "webhook ingest: "+reason)
	}

	_ = h.queries.MarkLatestLogBouncedByEmail(r.Context(), db.MarkLatestLogBouncedByEmailParams{
		ProjectID: project.ID,
		ToEmail:   email,
		Error:     sql.NullString{String: "bounce reported via webhook: " + reason, Valid: true},
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted", "email": email})
}

func extractBounce(body []byte) (string, string, bool) {
	if len(body) == 0 {
		return "", "", false
	}

	var generic genericBouncePayload
	if err := json.Unmarshal(body, &generic); err == nil && generic.Email != "" {
		reason := generic.Reason
		if reason == "" {
			reason = generic.Type
		}
		if reason == "" {
			reason = "external bounce report"
		}
		return strings.ToLower(strings.TrimSpace(generic.Email)), reason, true
	}

	var mg mailgunPayload
	if err := json.Unmarshal(body, &mg); err == nil && mg.EventData.Recipient != "" {
		if mg.EventData.Event != "failed" && mg.EventData.Event != "rejected" {
			return "", "", false
		}
		reason := mg.EventData.Reason
		if reason == "" {
			reason = "mailgun " + mg.EventData.Event
		}
		if mg.EventData.Severity != "" {
			reason = reason + " (" + mg.EventData.Severity + ")"
		}
		return strings.ToLower(strings.TrimSpace(mg.EventData.Recipient)), reason, true
	}

	return "", "", false
}

func writeBounceError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

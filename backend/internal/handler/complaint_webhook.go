package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/google/uuid"
)

const complaintWebhookMaxBody = 65536

type ComplaintWebhookHandler struct {
	queries      *db.Queries
	suppressions *service.SuppressionService
}

func NewComplaintWebhookHandler(queries *db.Queries, suppressions *service.SuppressionService) *ComplaintWebhookHandler {
	return &ComplaintWebhookHandler{queries: queries, suppressions: suppressions}
}

func (h *ComplaintWebhookHandler) Receive(w http.ResponseWriter, r *http.Request) {
	pid, err := uuid.Parse(r.PathValue("projectId"))
	if err != nil {
		writeBounceError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	tokenUUID, err := uuid.Parse(r.URL.Query().Get("token"))
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

	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, complaintWebhookMaxBody+1))
	if err != nil {
		writeBounceError(w, http.StatusBadRequest, "could not read body")
		return
	}
	if len(bodyBytes) > complaintWebhookMaxBody {
		writeBounceError(w, http.StatusRequestEntityTooLarge, "payload too large")
		return
	}

	email, reason, ok := extractComplaint(bodyBytes)
	if !ok {
		writeBounceError(w, http.StatusBadRequest, "could not find email in payload")
		return
	}

	if h.suppressions != nil {
		_, _ = h.suppressions.Add(r.Context(), project.ID, email, service.SuppressionReasonComplaint, "complaint webhook: "+reason)
	}

	_ = h.queries.MarkLatestLogComplainedByEmail(r.Context(), db.MarkLatestLogComplainedByEmailParams{
		ProjectID: project.ID,
		ToEmail:   email,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted", "email": email})
}

func extractComplaint(body []byte) (string, string, bool) {
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
			reason = "spam complaint"
		}
		return strings.ToLower(strings.TrimSpace(generic.Email)), reason, true
	}

	var mg mailgunPayload
	if err := json.Unmarshal(body, &mg); err == nil && mg.EventData.Recipient != "" {
		if mg.EventData.Event != "complained" {
			return "", "", false
		}
		return strings.ToLower(strings.TrimSpace(mg.EventData.Recipient)), "mailgun complained", true
	}

	return "", "", false
}

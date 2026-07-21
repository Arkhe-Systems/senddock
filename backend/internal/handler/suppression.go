package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/google/uuid"
)

type SuppressionHandler struct {
	suppressions   *service.SuppressionService
	projectService *service.ProjectService
	Audit          *service.AuditService
}

func NewSuppressionHandler(suppressions *service.SuppressionService, projectService *service.ProjectService) *SuppressionHandler {
	return &SuppressionHandler{suppressions: suppressions, projectService: projectService}
}

type suppressionResponse struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	Email      string `json:"email"`
	Reason     string `json:"reason"`
	Source     string `json:"source,omitempty"`
	CreatedAt  string `json:"created_at"`
	LastSeenAt string `json:"last_seen_at"`
}

func toSuppressionResponse(s db.Suppression) suppressionResponse {
	r := suppressionResponse{
		ID:         s.ID.String(),
		ProjectID:  s.ProjectID.String(),
		Email:      s.EmailNormalized,
		Reason:     s.Reason,
		CreatedAt:  s.CreatedAt.UTC().Format(time.RFC3339),
		LastSeenAt: s.LastSeenAt.UTC().Format(time.RFC3339),
	}
	if s.Source.Valid {
		r.Source = s.Source.String
	}
	return r
}

func (h *SuppressionHandler) verifyOwner(r *http.Request) (uuid.UUID, error) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	if _, err := h.projectService.GetByID(r.Context(), projectID, userID); err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(projectID)
}

func (h *SuppressionHandler) List(w http.ResponseWriter, r *http.Request) {
	pid, err := h.verifyOwner(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}

	limit := int32(50)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	reason := r.URL.Query().Get("reason")

	rows, total, err := h.suppressions.List(r.Context(), pid, reason, limit, offset)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list suppressions")
		return
	}

	out := make([]suppressionResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toSuppressionResponse(row))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"suppressions": out, "total": total})
}

type addSuppressionRequest struct {
	Emails []string `json:"emails"`
	Reason string   `json:"reason"`
	Source string   `json:"source"`
}

func (h *SuppressionHandler) Add(w http.ResponseWriter, r *http.Request) {
	projectID, _, ok := requireCap(w, r, h.projectService, service.CapSuppressionsWrite)
	if !ok {
		return
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req addSuppressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Emails) == 0 {
		writeJSONError(w, http.StatusBadRequest, "emails array is required")
		return
	}
	reason := req.Reason
	if reason == "" {
		reason = service.SuppressionReasonManual
	}

	added := 0
	skipped := 0
	for _, raw := range req.Emails {
		email := strings.ToLower(strings.TrimSpace(raw))
		if email == "" || !strings.Contains(email, "@") {
			skipped++
			continue
		}
		if _, err := h.suppressions.Add(r.Context(), pid, email, reason, req.Source); err != nil {
			skipped++
			continue
		}
		added++
	}

	if h.Audit != nil && added > 0 {
		userID := r.Context().Value(auth.UserIDKey).(string)
		h.Audit.LogFromRequest(r, pid.String(), userID, "suppression.add", "suppression", "", map[string]any{"added": added, "reason": reason})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"added": added, "skipped": skipped})
}

func (h *SuppressionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, _, ok := requireCap(w, r, h.projectService, service.CapSuppressionsWrite)
	if !ok {
		return
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	sid, err := uuid.Parse(r.PathValue("suppressionId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid suppression id")
		return
	}
	if err := h.suppressions.Remove(r.Context(), pid, sid); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to delete suppression")
		return
	}
	if h.Audit != nil {
		userID := r.Context().Value(auth.UserIDKey).(string)
		h.Audit.LogFromRequest(r, pid.String(), userID, "suppression.delete", "suppression", sid.String(), nil)
	}
	w.WriteHeader(http.StatusNoContent)
}


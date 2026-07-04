package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
)

var allowedWebhookEvents = map[string]bool{
	"email.sent":              true,
	"email.failed":            true,
	"email.bounced":           true,
	"email.opened":            true,
	"email.clicked":           true,
	"subscriber.created":      true,
	"subscriber.unsubscribed": true,
}

type WebhookHandler struct {
	webhookService *service.WebhookService
	projectService *service.ProjectService
	Audit          *service.AuditService
}

func NewWebhookHandler(webhookService *service.WebhookService, projectService *service.ProjectService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		projectService: projectService,
	}
}

type createWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type patchWebhookRequest struct {
	Active *bool `json:"active"`
}

func (h *WebhookHandler) verifyProjectAccess(r *http.Request) (string, string, error) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, userID, err
}

func validateWebhookURL(raw string) error {
	if raw == "" {
		return errors.New("url is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url must use http or https")
	}
	if u.Host == "" {
		return errors.New("url must include a host")
	}
	return nil
}

func normalizeWebhookEvents(events []string) ([]string, error) {
	if len(events) == 0 {
		out := make([]string, 0, len(allowedWebhookEvents))
		for ev := range allowedWebhookEvents {
			out = append(out, ev)
		}
		return out, nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(events))
	for _, ev := range events {
		ev = strings.TrimSpace(ev)
		if !allowedWebhookEvents[ev] {
			return nil, errors.New("unknown event: " + ev)
		}
		if !seen[ev] {
			seen[ev] = true
			out = append(out, ev)
		}
	}
	return out, nil
}

func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapWebhooksWrite)
	if !ok {
		return
	}

	var req createWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateWebhookURL(req.URL); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	events, err := normalizeWebhookEvents(req.Events)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	hook, err := h.webhookService.Create(r.Context(), projectID, req.URL, events)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create webhook")
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "webhook.create", "webhook", hook.ID.String(), map[string]any{"url": hook.Url})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.FromWebhook(hook))
}

func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectAccess(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}

	hooks, err := h.webhookService.List(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list webhooks")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"webhooks": response.FromWebhooks(hooks)})
}

func (h *WebhookHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectAccess(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}

	hook, err := h.webhookService.Get(r.Context(), projectID, r.PathValue("webhookId"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "webhook not found")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromWebhook(hook))
}

func (h *WebhookHandler) Patch(w http.ResponseWriter, r *http.Request) {
	projectID, _, ok := requireCap(w, r, h.projectService, service.CapWebhooksWrite)
	if !ok {
		return
	}

	var req patchWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Active == nil {
		writeJSONError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	hook, err := h.webhookService.UpdateActive(r.Context(), projectID, r.PathValue("webhookId"), *req.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "webhook not found")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromWebhook(hook))
}

func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapWebhooksWrite)
	if !ok {
		return
	}

	webhookID := r.PathValue("webhookId")
	if err := h.webhookService.Delete(r.Context(), projectID, webhookID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "webhook.delete", "webhook", webhookID, nil)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WebhookHandler) Deliveries(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectAccess(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}

	limit := int32(50)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = int32(n)
		}
	}

	deliveries, err := h.webhookService.ListDeliveries(r.Context(), projectID, r.PathValue("webhookId"), limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "webhook not found")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"deliveries": response.FromWebhookDeliveries(deliveries)})
}

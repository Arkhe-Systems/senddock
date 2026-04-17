package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/arkhe-systems/senddock/internal/middleware"
	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
)

type EmailHandler struct {
	emailService   *service.EmailService
	projectService *service.ProjectService
}

func NewEmailHandler(emailService *service.EmailService, projectService *service.ProjectService) *EmailHandler {
	return &EmailHandler{
		emailService:   emailService,
		projectService: projectService,
	}
}

type sendToSubscriberRequest struct {
	SubscriberID string `json:"subscriber_id"`
	TemplateID   string `json:"template_id"`
}

type broadcastRequest struct {
	TemplateID string `json:"template_id"`
}

type sendDirectRequest struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	HtmlBody string `json:"html_body"`
}

type sendTemplateRequest struct {
	TemplateID string            `json:"template_id"`
	To         string            `json:"to"`
	Data       map[string]string `json:"data"`
}

func (h *EmailHandler) verifyAccess(r *http.Request) (string, error) {
	if pid, ok := r.Context().Value(middleware.ProjectIDKey).(string); ok {
		return pid, nil
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, err
}

func (h *EmailHandler) SendToSubscriber(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req sendToSubscriberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.SubscriberID == "" || req.TemplateID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "subscriber_id and template_id are required"})
		return
	}

	result, err := h.emailService.SendToSubscriber(r.Context(), projectID, req.SubscriberID, req.TemplateID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *EmailHandler) Broadcast(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req broadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.TemplateID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "template_id is required"})
		return
	}

	result, err := h.emailService.Broadcast(r.Context(), projectID, req.TemplateID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *EmailHandler) SendDirect(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req sendDirectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.To == "" || req.Subject == "" || req.HtmlBody == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "to, subject, and html_body are required"})
		return
	}

	err = h.emailService.SendDirect(r.Context(), projectID, req.To, req.Subject, req.HtmlBody)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "sent"})
}

func (h *EmailHandler) SendTemplate(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req sendTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.TemplateID == "" || req.To == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "template_id and to are required"})
		return
	}

	err = h.emailService.SendWithTemplate(r.Context(), projectID, req.TemplateID, req.To, req.Data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "sent"})
}

func (h *EmailHandler) TestSMTP(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	err = h.emailService.TestSMTP(r.Context(), projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "SMTP connection successful. Test email sent."})
}

func (h *EmailHandler) Logs(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	limit := int32(50)
	offset := int32(0)

	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		limit = int32(v)
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = int32(v)
	}

	logs, total, err := h.emailService.GetLogs(r.Context(), projectID, limit, offset)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  response.FromEmailLogs(logs),
		"total": total,
	})
}

func (h *EmailHandler) Stats(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	stats, err := h.emailService.GetStats(r.Context(), projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

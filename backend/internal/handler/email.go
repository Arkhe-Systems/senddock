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

type sendRequest struct {
	To           string            `json:"to"`
	SubscriberID string            `json:"subscriber_id"`
	TemplateID   string            `json:"template_id"`
	Subject      string            `json:"subject"`
	HtmlBody     string            `json:"html_body"`
	Data         map[string]string `json:"data"`
}

type broadcastRequest struct {
	TemplateID string `json:"template_id"`
}

type batchRecipient struct {
	To   string            `json:"to"`
	Data map[string]string `json:"data"`
}

type batchSendRequest struct {
	TemplateID string           `json:"template_id"`
	Subject    string           `json:"subject"`
	Recipients []batchRecipient `json:"recipients"`
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

func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if req.SubscriberID != "" && req.TemplateID != "" {
		result, err := h.emailService.SendToSubscriber(r.Context(), projectID, req.SubscriberID, req.TemplateID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	if req.To != "" && req.TemplateID != "" {
		err := h.emailService.SendWithTemplate(r.Context(), projectID, req.TemplateID, req.To, req.Subject, req.Data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "sent"})
		return
	}

	if req.To != "" && req.HtmlBody != "" {
		if req.Subject == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: "subject is required for direct send"})
			return
		}
		err := h.emailService.SendDirect(r.Context(), projectID, req.To, req.Subject, req.HtmlBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "sent"})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorResponse{Error: "provide (to + template_id), (subscriber_id + template_id), or (to + subject + html_body)"})
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

func (h *EmailHandler) BatchSend(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyAccess(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req batchSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.TemplateID == "" || len(req.Recipients) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "template_id and recipients are required"})
		return
	}

	sent := 0
	failed := 0
	for _, rcpt := range req.Recipients {
		if rcpt.To == "" {
			failed++
			continue
		}
		err := h.emailService.SendWithTemplate(r.Context(), projectID, req.TemplateID, rcpt.To, req.Subject, rcpt.Data)
		if err != nil {
			failed++
		} else {
			sent++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"sent": sent, "failed": failed})
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

	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 100 {
		limit = int32(v)
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = int32(v)
	}

	status := r.URL.Query().Get("status")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	logs, total, err := h.emailService.GetLogs(r.Context(), projectID, limit, offset, status, from, to)
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

func (h *EmailHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	subscriberID := r.PathValue("subscriberId")

	err := h.emailService.Unsubscribe(r.Context(), projectID, subscriberID)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html><body style='font-family:sans-serif;text-align:center;padding:60px'><h2>Link expired or invalid</h2></body></html>"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<html><body style='font-family:sans-serif;text-align:center;padding:60px'><h2>You have been unsubscribed</h2><p>You will no longer receive emails from this project.</p></body></html>"))
}

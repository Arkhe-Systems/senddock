package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/service"
)

type WaitlistHandler struct {
	subscriberService *service.SubscriberService
	emailService      *service.EmailService
	projectID         string
	templateID        string
}

func NewWaitlistHandler(subscriberService *service.SubscriberService, emailService *service.EmailService, projectID, templateID string) *WaitlistHandler {
	return &WaitlistHandler{
		subscriberService: subscriberService,
		emailService:      emailService,
		projectID:         projectID,
		templateID:        templateID,
	}
}

type waitlistRequest struct {
	Email string `json:"email"`
}

func (h *WaitlistHandler) Join(w http.ResponseWriter, r *http.Request) {
	if h.projectID == "" || h.templateID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(errorResponse{Error: "waitlist not configured"})
		return
	}

	var req waitlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" || !isValidEmail(req.Email) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "valid email is required"})
		return
	}

	_, err := h.subscriberService.Create(r.Context(), h.projectID, req.Email, "", "pending")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorResponse{Error: "already on the waitlist"})
		return
	}

	go h.emailService.SendWithTemplate(
		r.Context(),
		h.projectID,
		h.templateID,
		req.Email,
		"",
		map[string]string{"email": req.Email},
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "joined"})
}

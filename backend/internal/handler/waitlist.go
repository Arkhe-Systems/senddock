package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/service"
)

type WaitlistHandler struct {
	subscriberService *service.SubscriberService
	emailService      *service.EmailService
}

func NewWaitlistHandler(subscriberService *service.SubscriberService, emailService *service.EmailService) *WaitlistHandler {
	return &WaitlistHandler{
		subscriberService: subscriberService,
		emailService:      emailService,
	}
}

type waitlistJoinRequest struct {
	Email      string `json:"email"`
	TemplateID string `json:"template_id"`
}

func (h *WaitlistHandler) Join(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	projectID := r.PathValue("id")

	var req waitlistJoinRequest
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

	_, err := h.subscriberService.Create(r.Context(), projectID, req.Email, "", "pending")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorResponse{Error: "already on the waitlist"})
		return
	}

	if req.TemplateID != "" {
		go func() {
			err := h.emailService.SendWithTemplate(
				context.Background(),
				projectID,
				req.TemplateID,
				req.Email,
				"",
				map[string]string{"email": req.Email},
			)
			if err != nil {
				log.Printf("Waitlist confirmation email failed for %s: %v", req.Email, err)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "joined"})
}

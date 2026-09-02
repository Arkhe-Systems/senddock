package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
)

type NewsletterHandler struct {
	newsletterService *service.NewsletterService
	projectService    *service.ProjectService
	Audit             *service.AuditService
}

func NewNewsletterHandler(newsletterService *service.NewsletterService, projectService *service.ProjectService) *NewsletterHandler {
	return &NewsletterHandler{
		newsletterService: newsletterService,
		projectService:    projectService,
	}
}

type newsletterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type setNewslettersRequest struct {
	NewsletterIDs []string `json:"newsletter_ids"`
}

func (h *NewsletterHandler) verifyProjectAccess(r *http.Request) (string, string, error) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, userID, err
}

func (h *NewsletterHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectAccess(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}

	newsletters, err := h.newsletterService.List(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromNewsletters(newsletters))
}

func (h *NewsletterHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapSubscribersWrite)
	if !ok {
		return
	}

	var req newsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newsletter, err := h.newsletterService.Create(r.Context(), projectID, req.Name, req.Description)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "newsletter.created", "newsletter", newsletter.ID.String(), map[string]any{"name": newsletter.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.FromNewsletter(newsletter, 0))
}

func (h *NewsletterHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapSubscribersWrite)
	if !ok {
		return
	}

	var req newsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newsletter, err := h.newsletterService.Update(r.Context(), r.PathValue("newsletterId"), projectID, req.Name, req.Description)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "newsletter.updated", "newsletter", newsletter.ID.String(), map[string]any{"name": newsletter.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromNewsletter(newsletter, 0))
}

func (h *NewsletterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapSubscribersWrite)
	if !ok {
		return
	}

	newsletterID := r.PathValue("newsletterId")
	if err := h.newsletterService.Delete(r.Context(), newsletterID, projectID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "newsletter.deleted", "newsletter", newsletterID, nil)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NewsletterHandler) ListForSubscriber(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectAccess(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}

	memberships, err := h.newsletterService.ListForSubscriber(r.Context(), projectID, r.PathValue("subscriberId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromSubscriberNewsletters(memberships))
}

func (h *NewsletterHandler) SetForSubscriber(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapSubscribersWrite)
	if !ok {
		return
	}

	var req setNewslettersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	subscriberID := r.PathValue("subscriberId")
	if err := h.newsletterService.SetSubscriberNewsletters(r.Context(), projectID, subscriberID, req.NewsletterIDs); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "subscriber.newsletters_updated", "subscriber", subscriberID, map[string]any{"count": len(req.NewsletterIDs)})
	}

	memberships, err := h.newsletterService.ListForSubscriber(r.Context(), projectID, subscriberID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromSubscriberNewsletters(memberships))
}

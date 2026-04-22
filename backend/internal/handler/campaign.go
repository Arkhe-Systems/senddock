package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/arkhe-systems/senddock/internal/middleware"
	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
)

type CampaignHandler struct {
	campaignService *service.CampaignService
	projectService  *service.ProjectService
}

func NewCampaignHandler(campaignService *service.CampaignService, projectService *service.ProjectService) *CampaignHandler {
	return &CampaignHandler{
		campaignService: campaignService,
		projectService:  projectService,
	}
}

type createCampaignRequest struct {
	TemplateID  string            `json:"template_id"`
	Name        string            `json:"name"`
	ScheduledAt string            `json:"scheduled_at"`
	Variables   map[string]string `json:"variables"`
}

func (h *CampaignHandler) verifyProjectOwner(r *http.Request) (string, error) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, err
}

func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req createCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" || req.TemplateID == "" || req.ScheduledAt == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "name, template_id, and scheduled_at are required"})
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "scheduled_at must be RFC3339 format"})
		return
	}

	var variablesJson []byte
	if req.Variables != nil {
		variablesJson, _ = json.Marshal(req.Variables)
	} else {
		variablesJson = []byte("{}")
	}

	campaign, err := h.campaignService.Create(r.Context(), projectID, req.TemplateID, req.Name, scheduledAt, variablesJson)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.FromCampaign(campaign))
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	campaigns, err := h.campaignService.ListByProject(r.Context(), projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "failed to list campaigns"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromCampaigns(campaigns))
}

func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	campaignID := r.PathValue("campaignId")

	err = h.campaignService.Delete(r.Context(), campaignID, projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "campaign not found or already sent"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	campaignID := r.PathValue("campaignId")

	var req createCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" || req.TemplateID == "" || req.ScheduledAt == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "name, template_id, and scheduled_at are required"})
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "scheduled_at must be RFC3339 format"})
		return
	}

	var variablesJson []byte
	if req.Variables != nil {
		variablesJson, _ = json.Marshal(req.Variables)
	} else {
		variablesJson = []byte("{}")
	}

	campaign, err := h.campaignService.Update(r.Context(), campaignID, projectID, req.TemplateID, req.Name, scheduledAt, variablesJson)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response.FromCampaign(campaign))
}

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/middleware"
	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
)

type TemplateHandler struct {
	templateService *service.TemplateService
	projectService  *service.ProjectService
}

func NewTemplateHandler(templateService *service.TemplateService, projectService *service.ProjectService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
		projectService:  projectService,
	}
}

type createTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	HtmlBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

type updateTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	HtmlBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

func (h *TemplateHandler) verifyProjectOwner(r *http.Request) (string, error) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, err
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req createTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "name is required"})
		return
	}

	template, err := h.templateService.Create(r.Context(), projectID, req.Name, req.Subject, req.HtmlBody, req.TextBody)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.FromTemplate(template))
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	templates, err := h.templateService.ListByProject(r.Context(), projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromTemplates(templates))
}

func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	templateID := r.PathValue("templateId")

	template, err := h.templateService.GetByID(r.Context(), templateID, projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "template not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromTemplate(template))
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	templateID := r.PathValue("templateId")

	var req updateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "name is required"})
		return
	}

	template, err := h.templateService.Update(r.Context(), templateID, projectID, req.Name, req.Subject, req.HtmlBody, req.TextBody)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "template not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromTemplate(template))
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	templateID := r.PathValue("templateId")

	err = h.templateService.Delete(r.Context(), templateID, projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "template not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

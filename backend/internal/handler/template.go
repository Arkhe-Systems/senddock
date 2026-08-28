package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
)

type TemplateHandler struct {
	templateService *service.TemplateService
	projectService  *service.ProjectService
	libraryService  *service.TemplateLibraryService
}

func NewTemplateHandler(templateService *service.TemplateService, projectService *service.ProjectService, libraryService *service.TemplateLibraryService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
		projectService:  projectService,
		libraryService:  libraryService,
	}
}

type createTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	HtmlBody string `json:"html_body"`
	TextBody string `json:"text_body"`
	Type     string `json:"type"`
}

type updateTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	HtmlBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

func (h *TemplateHandler) verifyProjectOwner(r *http.Request) (string, error) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, err
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, _, ok := requireCap(w, r, h.projectService, service.CapTemplatesWrite)
	if !ok {
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

	templateType := req.Type
	if templateType == "" {
		templateType = "email"
	}
	if templateType != "email" && templateType != "page" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "type must be email or page"})
		return
	}

	template, err := h.templateService.Create(r.Context(), projectID, req.Name, req.Subject, req.HtmlBody, req.TextBody, templateType)
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
	projectID, _, ok := requireCap(w, r, h.projectService, service.CapTemplatesWrite)
	if !ok {
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
	projectID, _, ok := requireCap(w, r, h.projectService, service.CapTemplatesWrite)
	if !ok {
		return
	}

	templateID := r.PathValue("templateId")

	err := h.templateService.Delete(r.Context(), templateID, projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "template not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TemplateHandler) LibraryList(w http.ResponseWriter, r *http.Request) {
	if _, err := h.verifyProjectOwner(r); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	entries, err := h.libraryService.List(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(errorResponse{Error: "template library is unavailable"})
		return
	}

	if entries == nil {
		entries = []service.TemplateLibraryEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *TemplateHandler) LibraryUse(w http.ResponseWriter, r *http.Request) {
	projectID, _, ok := requireCap(w, r, h.projectService, service.CapTemplatesWrite)
	if !ok {
		return
	}

	libraryID := r.PathValue("libraryId")
	if libraryID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "library id is required"})
		return
	}

	entry, err := h.libraryService.Find(r.Context(), libraryID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrTemplateLibraryNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse{Error: "template not found in library"})
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(errorResponse{Error: "template library is unavailable"})
		return
	}

	html, err := h.libraryService.FetchHTML(r.Context(), entry)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(errorResponse{Error: "failed to fetch template from library"})
		return
	}

	template, err := h.templateService.Create(r.Context(), projectID, entry.Name, "", html, "", "email")
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

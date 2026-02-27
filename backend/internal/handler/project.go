package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/middleware"
	"github.com/arkhe-systems/senddock/internal/service"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

type createProjectRequest struct {
	Name      string `json:"name"`
	FromName  string `json:"from_name"`
	FromEmail string `json:"from_email"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" || req.FromEmail == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "name and from_email are required"})
		return
	}

	project, err := h.projectService.Create(r.Context(), userID, req.Name, req.FromName, req.FromEmail)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	projects, err := h.projectService.ListByUser(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := r.PathValue("id")

	project, err := h.projectService.GetByID(r.Context(), projectID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := r.PathValue("id")

	err := h.projectService.Delete(r.Context(), projectID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

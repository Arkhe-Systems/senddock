package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
)

type APIKeyHandler struct {
	apiKeyService  *service.APIKeyService
	projectService *service.ProjectService
	Audit          *service.AuditService
}

func NewAPIKeyHandler(apiKeyService *service.APIKeyService, projectService *service.ProjectService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService:  apiKeyService,
		projectService: projectService,
	}
}

type createAPIKeyRequest struct {
	Name string `json:"name"`
}

func (h *APIKeyHandler) verifyProjectOwner(r *http.Request) (string, error) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, err
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req createAPIKeyRequest
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

	result, err := h.apiKeyService.Create(r.Context(), projectID, req.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	if h.Audit != nil {
		userID := r.Context().Value(auth.UserIDKey).(string)
		h.Audit.LogFromRequest(r, projectID, userID, "api_key.create", "api_key", result.APIKey.ID.String(), map[string]any{"name": result.APIKey.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":     result.Key,
		"api_key": response.FromAPIKey(result.APIKey),
	})
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	keys, err := h.apiKeyService.ListByProject(r.Context(), projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromAPIKeys(keys))
}

func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	keyID := r.PathValue("keyId")

	err = h.apiKeyService.Delete(r.Context(), keyID, projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "api key not found"})
		return
	}

	if h.Audit != nil {
		userID := r.Context().Value(auth.UserIDKey).(string)
		h.Audit.LogFromRequest(r, projectID, userID, "api_key.revoke", "api_key", keyID, nil)
	}

	w.WriteHeader(http.StatusNoContent)
}

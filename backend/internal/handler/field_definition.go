package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
)

type FieldDefinitionHandler struct {
	fieldService   *service.FieldDefinitionService
	projectService *service.ProjectService
	Audit          *service.AuditService
}

func NewFieldDefinitionHandler(fieldService *service.FieldDefinitionService, projectService *service.ProjectService) *FieldDefinitionHandler {
	return &FieldDefinitionHandler{
		fieldService:   fieldService,
		projectService: projectService,
	}
}

type createFieldRequest struct {
	Key       string   `json:"key"`
	Label     string   `json:"label"`
	FieldType string   `json:"field_type"`
	Options   []string `json:"options"`
	Required  bool     `json:"required"`
}

type updateFieldRequest struct {
	Label    string   `json:"label"`
	Options  []string `json:"options"`
	Required bool     `json:"required"`
}

func (h *FieldDefinitionHandler) verifyProjectAccess(r *http.Request) (string, string, error) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, userID, err
}

func (h *FieldDefinitionHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectAccess(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}

	defs, err := h.fieldService.List(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromFieldDefinitions(defs))
}

func (h *FieldDefinitionHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapProjectSettings)
	if !ok {
		return
	}

	var req createFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	def, err := h.fieldService.Create(r.Context(), projectID, service.FieldDefinitionInput{
		Key:      req.Key,
		Label:    req.Label,
		Type:     req.FieldType,
		Options:  req.Options,
		Required: req.Required,
	})
	if err != nil {
		h.writeCreateError(w, err)
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "field.created", "field", def.ID.String(), map[string]any{"key": def.Key})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.FromFieldDefinition(def))
}

func (h *FieldDefinitionHandler) writeCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidFieldKey), errors.Is(err, service.ErrInvalidFieldType), errors.Is(err, service.ErrEnumNeedsOptions):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	case strings.Contains(err.Error(), "duplicate key"), strings.Contains(err.Error(), "unique"):
		writeJSONError(w, http.StatusConflict, "a field with this key already exists")
	default:
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *FieldDefinitionHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapProjectSettings)
	if !ok {
		return
	}

	fieldID := r.PathValue("fieldId")

	var req updateFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	def, err := h.fieldService.Update(r.Context(), projectID, fieldID, req.Label, req.Options, req.Required)
	if err != nil {
		if errors.Is(err, service.ErrEnumNeedsOptions) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusNotFound, "field not found")
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "field.updated", "field", fieldID, map[string]any{"key": def.Key})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromFieldDefinition(def))
}

func (h *FieldDefinitionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapProjectSettings)
	if !ok {
		return
	}

	fieldID := r.PathValue("fieldId")

	if err := h.fieldService.Delete(r.Context(), projectID, fieldID); err != nil {
		writeJSONError(w, http.StatusNotFound, "field not found")
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "field.deleted", "field", fieldID, nil)
	}

	w.WriteHeader(http.StatusNoContent)
}

package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
)

type SegmentHandler struct {
	segmentService *service.SegmentService
	projectService *service.ProjectService
	Audit          *service.AuditService
}

func NewSegmentHandler(segmentService *service.SegmentService, projectService *service.ProjectService) *SegmentHandler {
	return &SegmentHandler{
		segmentService: segmentService,
		projectService: projectService,
	}
}

type segmentRequest struct {
	Name      string          `json:"name"`
	Predicate json.RawMessage `json:"predicate"`
}

type previewRequest struct {
	Predicate json.RawMessage `json:"predicate"`
}

func (h *SegmentHandler) verifyProjectAccess(r *http.Request) (string, string, error) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, userID, err
}

func (h *SegmentHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectAccess(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}
	segments, err := h.segmentService.List(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromSegments(segments))
}

func (h *SegmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapSubscribersWrite)
	if !ok {
		return
	}
	var req segmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	segment, err := h.segmentService.Create(r.Context(), projectID, req.Name, req.Predicate)
	if err != nil {
		h.writeSaveError(w, err)
		return
	}
	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "segment.created", "segment", segment.ID.String(), map[string]any{"name": segment.Name})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.FromSegment(segment))
}

func (h *SegmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapSubscribersWrite)
	if !ok {
		return
	}
	segmentID := r.PathValue("segmentId")
	var req segmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	segment, err := h.segmentService.Update(r.Context(), projectID, segmentID, req.Name, req.Predicate)
	if err != nil {
		h.writeSaveError(w, err)
		return
	}
	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "segment.updated", "segment", segmentID, map[string]any{"name": segment.Name})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromSegment(segment))
}

func (h *SegmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapSubscribersWrite)
	if !ok {
		return
	}
	segmentID := r.PathValue("segmentId")
	if err := h.segmentService.Delete(r.Context(), projectID, segmentID); err != nil {
		writeJSONError(w, http.StatusNotFound, "segment not found")
		return
	}
	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "segment.deleted", "segment", segmentID, nil)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SegmentHandler) Preview(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectAccess(r)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}
	var req previewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	count, err := h.segmentService.Preview(r.Context(), projectID, req.Predicate, true)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPredicate) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *SegmentHandler) writeSaveError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrInvalidPredicate) {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

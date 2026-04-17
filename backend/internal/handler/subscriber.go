package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/arkhe-systems/senddock/internal/middleware"
	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
)

type SubscriberHandler struct {
	subscriberService *service.SubscriberService
	projectService    *service.ProjectService
}

func NewSubscriberHandler(subscriberService *service.SubscriberService, projectService *service.ProjectService) *SubscriberHandler {
	return &SubscriberHandler{
		subscriberService: subscriberService,
		projectService:    projectService,
	}
}

type createSubscriberRequest struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *SubscriberHandler) verifyProjectOwner(r *http.Request) (string, string, error) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	projectID := r.PathValue("id")

	_, err := h.projectService.GetByID(r.Context(), projectID, userID)
	return projectID, userID, err
}

func (h *SubscriberHandler) Import(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var subscribers []service.ImportSubscriber
	if err := json.NewDecoder(r.Body).Decode(&subscribers); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body, expected array of subscribers"})
		return
	}

	if len(subscribers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "empty array"})
		return
	}

	result, err := h.subscriberService.BulkImport(r.Context(), projectID, subscribers)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *SubscriberHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	var req createSubscriberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "email is required"})
		return
	}

	subscriber, err := h.subscriberService.Create(r.Context(), projectID, req.Email, req.Name, req.Status)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorResponse{Error: "subscriber already exists"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.FromSubscriber(subscriber))
}

func (h *SubscriberHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int32(50)
	offset := int32(0)

	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = int32(v)
		}
	}

	subscribers, err := h.subscriberService.ListByProject(r.Context(), projectID, limit, offset)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	count, _ := h.subscriberService.CountByProject(r.Context(), projectID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscribers": response.FromSubscribers(subscribers),
		"total":       count,
	})
}

func (h *SubscriberHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	subscriberID := r.PathValue("subscriberId")

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Status != "active" && req.Status != "unsubscribed" && req.Status != "pending" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "status must be active, pending, or unsubscribed"})
		return
	}

	subscriber, err := h.subscriberService.UpdateStatus(r.Context(), subscriberID, projectID, req.Status)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "subscriber not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromSubscriber(subscriber))
}

func (h *SubscriberHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, _, err := h.verifyProjectOwner(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	subscriberID := r.PathValue("subscriberId")

	err = h.subscriberService.Delete(r.Context(), subscriberID, projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "subscriber not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

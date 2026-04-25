package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/service"
)

type ReleaseHandler struct {
	releaseService *service.ReleaseService
}

func NewReleaseHandler(releaseService *service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{releaseService: releaseService}
}

func (h *ReleaseHandler) Get(w http.ResponseWriter, r *http.Request) {
	info := h.releaseService.GetRelease(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

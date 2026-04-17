package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

var transparentPixel, _ = base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")

type TrackingHandler struct {
	queries *db.Queries
}

func NewTrackingHandler(queries *db.Queries) *TrackingHandler {
	return &TrackingHandler{queries: queries}
}

func (h *TrackingHandler) Open(w http.ResponseWriter, r *http.Request) {
	logID := r.PathValue("logId")

	if lid, err := uuid.Parse(logID); err == nil {
		h.queries.MarkEmailOpened(r.Context(), lid)
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Write(transparentPixel)
}

package handler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/internal/webhooks"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

var transparentPixel, _ = base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")

type TrackingHandler struct {
	queries *db.Queries
	emails  *service.EmailService
	hooks   *webhooks.Service
}

func NewTrackingHandler(queries *db.Queries, emails *service.EmailService, hooks *webhooks.Service) *TrackingHandler {
	return &TrackingHandler{queries: queries, emails: emails, hooks: hooks}
}

func (h *TrackingHandler) Open(w http.ResponseWriter, r *http.Request) {
	logID := strings.TrimSuffix(r.PathValue("logId"), ".gif")

	if lid, err := uuid.Parse(logID); err == nil {
		rows, err := h.queries.MarkEmailOpened(r.Context(), lid)
		if err == nil && rows > 0 {
			h.dispatchEmailEvent(r.Context(), "email.opened", lid, map[string]any{"opened_at": time.Now().UTC().Format(time.RFC3339)})
		}
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Write(transparentPixel)
}

func (h *TrackingHandler) dispatchEmailEvent(ctx context.Context, eventType string, logID uuid.UUID, extra map[string]any) {
	if h.hooks == nil {
		return
	}
	projectID, err := h.queries.GetEmailLogProjectID(ctx, logID)
	if err != nil {
		return
	}
	data := map[string]any{
		"log_id":     logID.String(),
		"project_id": projectID.String(),
	}
	for k, v := range extra {
		data[k] = v
	}
	h.hooks.Enqueue(ctx, webhooks.Event{
		Type:      eventType,
		ProjectID: projectID,
		Data:      data,
	})
}

func (h *TrackingHandler) Click(w http.ResponseWriter, r *http.Request) {
	logIDStr := r.PathValue("logId")
	payload := r.PathValue("payload")

	dot := strings.LastIndex(payload, ".")
	if dot <= 0 {
		http.Error(w, "invalid tracking link", http.StatusBadRequest)
		return
	}
	encodedURL := payload[:dot]
	token := payload[dot+1:]

	rawURL, err := h.emails.DecodeClickURL(encodedURL)
	if err != nil {
		http.Error(w, "invalid tracking link", http.StatusBadRequest)
		return
	}

	if !h.emails.VerifyClickToken(logIDStr, rawURL, token) {
		http.Error(w, "invalid tracking link", http.StatusBadRequest)
		return
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		http.Error(w, "invalid tracking link", http.StatusBadRequest)
		return
	}

	if lid, err := uuid.Parse(logIDStr); err == nil {
		rows, _ := h.queries.MarkEmailClicked(r.Context(), lid)
		h.queries.CreateEmailClick(r.Context(), db.CreateEmailClickParams{
			LogID:     lid,
			Url:       rawURL,
			UrlHash:   shortHash(rawURL),
			UserAgent: nullString(r.UserAgent()),
			IpAddress: parseIP(r),
		})
		if rows > 0 {
			h.dispatchEmailEvent(r.Context(), "email.clicked", lid, map[string]any{
				"url":        rawURL,
				"clicked_at": time.Now().UTC().Format(time.RFC3339),
			})
		}
	}

	http.Redirect(w, r, rawURL, http.StatusFound)
}

func shortHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(h[:])[:16]
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func parseIP(r *http.Request) pqtype.Inet {
	host := r.Header.Get("X-Forwarded-For")
	if comma := strings.Index(host, ","); comma >= 0 {
		host = host[:comma]
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	if host == "" {
		return pqtype.Inet{}
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return pqtype.Inet{}
	}
	mask := net.CIDRMask(32, 32)
	if ip.To4() == nil {
		mask = net.CIDRMask(128, 128)
	}
	return pqtype.Inet{IPNet: net.IPNet{IP: ip, Mask: mask}, Valid: true}
}

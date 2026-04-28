package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

var transparentPixel, _ = base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")

type TrackingHandler struct {
	queries *db.Queries
	emails  *service.EmailService
}

func NewTrackingHandler(queries *db.Queries, emails *service.EmailService) *TrackingHandler {
	return &TrackingHandler{queries: queries, emails: emails}
}

func (h *TrackingHandler) Open(w http.ResponseWriter, r *http.Request) {
	logID := strings.TrimSuffix(r.PathValue("logId"), ".gif")

	if lid, err := uuid.Parse(logID); err == nil {
		h.queries.MarkEmailOpened(r.Context(), lid)
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Write(transparentPixel)
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
		h.queries.MarkEmailClicked(r.Context(), lid)
		h.queries.CreateEmailClick(r.Context(), db.CreateEmailClickParams{
			LogID:     lid,
			Url:       rawURL,
			UrlHash:   shortHash(rawURL),
			UserAgent: nullString(r.UserAgent()),
			IpAddress: parseIP(r),
		})
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

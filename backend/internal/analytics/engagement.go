package analytics

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Breakdown struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type Funnel struct {
	Sent    int64 `json:"sent"`
	Opened  int64 `json:"opened"`
	Clicked int64 `json:"clicked"`
}

type EngagementBucket struct {
	Bucket string `json:"bucket"`
	Opens  int64  `json:"opens"`
	Clicks int64  `json:"clicks"`
}

type HeatCell struct {
	Weekday int   `json:"weekday"`
	Hour    int   `json:"hour"`
	Count   int64 `json:"count"`
}

type Engagement struct {
	Devices []Breakdown        `json:"devices"`
	Clients []Breakdown        `json:"clients"`
	Funnel  Funnel             `json:"funnel"`
	Series  []EngagementBucket `json:"series"`
	Heatmap []HeatCell         `json:"heatmap"`
}

func (h *Handler) Engagement(w http.ResponseWriter, r *http.Request) {
	projectID, ok := h.authorizedProjectFromRequest(w, r)
	if !ok {
		return
	}
	from, to, granularity, ok := parseWindow(w, r)
	if !ok {
		return
	}

	nlFilter, nlAliased := "", ""
	nlArgs := []any{}
	if newsletterID := strings.TrimSpace(r.URL.Query().Get("newsletter_id")); newsletterID != "" {
		if _, err := uuid.Parse(newsletterID); err != nil {
			writeError(w, http.StatusBadRequest, "invalid newsletter_id")
			return
		}
		nlFilter = " AND newsletter_id = $4"
		nlAliased = " AND l.newsletter_id = $4"
		nlArgs = append(nlArgs, newsletterID)
	}

	out := Engagement{
		Devices: []Breakdown{},
		Clients: []Breakdown{},
		Series:  []EngagementBucket{},
		Heatmap: []HeatCell{},
	}

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT COALESCE(ec.user_agent, '')
		FROM email_clicks ec
		JOIN email_logs l ON l.id = ec.log_id
		WHERE l.project_id = $1 AND ec.clicked_at >= $2 AND ec.clicked_at < $3`+nlAliased+`
	`, append([]any{projectID, from, to}, nlArgs...)...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load engagement")
		return
	}
	defer rows.Close()

	devices := map[string]int64{}
	clients := map[string]int64{}
	for rows.Next() {
		var ua string
		if err := rows.Scan(&ua); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read engagement")
			return
		}
		devices[deviceClass(ua)]++
		clients[mailClient(ua)]++
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read engagement")
		return
	}
	out.Devices = sortedBreakdown(devices)
	out.Clients = sortedBreakdown(clients)

	if err := h.db.QueryRowContext(r.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE status = 'sent'),
			COUNT(*) FILTER (WHERE opened_at IS NOT NULL),
			COUNT(*) FILTER (WHERE clicked_at IS NOT NULL)
		FROM email_logs
		WHERE project_id = $1 AND sent_at >= $2 AND sent_at < $3`+nlFilter+`
	`, append([]any{projectID, from, to}, nlArgs...)...).Scan(&out.Funnel.Sent, &out.Funnel.Opened, &out.Funnel.Clicked); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute funnel")
		return
	}

	series, err := h.engagementSeries(r.Context(), projectID, from, to, granularity, nlFilter, nlArgs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute engagement series")
		return
	}
	out.Series = series

	heatmap, err := h.clickHeatmap(r.Context(), projectID, from, to, nlFilter, nlArgs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute heatmap")
		return
	}
	out.Heatmap = heatmap

	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) engagementSeries(ctx context.Context, projectID string, from, to time.Time, granularity, nlFilter string, nlArgs []any) ([]EngagementBucket, error) {
	byBucket := map[string]*EngagementBucket{}

	tally := func(column string, assign func(*EngagementBucket, int64)) error {
		rows, err := h.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT to_char(date_trunc('%s', %s) AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS bucket, COUNT(*)
			FROM email_logs
			WHERE project_id = $1 AND %s IS NOT NULL AND %s >= $2 AND %s < $3` + nlFilter + `
			GROUP BY bucket
		`, granularity, column, column, column, column), append([]any{projectID, from, to}, nlArgs...)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var bucket string
			var count int64
			if err := rows.Scan(&bucket, &count); err != nil {
				return err
			}
			b := byBucket[bucket]
			if b == nil {
				b = &EngagementBucket{Bucket: bucket}
				byBucket[bucket] = b
			}
			assign(b, count)
		}
		return rows.Err()
	}

	if err := tally("opened_at", func(b *EngagementBucket, n int64) { b.Opens = n }); err != nil {
		return nil, err
	}
	if err := tally("clicked_at", func(b *EngagementBucket, n int64) { b.Clicks = n }); err != nil {
		return nil, err
	}

	buckets := make([]string, 0, len(byBucket))
	for k := range byBucket {
		buckets = append(buckets, k)
	}
	sort.Strings(buckets)

	out := make([]EngagementBucket, 0, len(buckets))
	for _, k := range buckets {
		out = append(out, *byBucket[k])
	}
	return out, nil
}

func (h *Handler) clickHeatmap(ctx context.Context, projectID string, from, to time.Time, nlFilter string, nlArgs []any) ([]HeatCell, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT EXTRACT(DOW FROM clicked_at AT TIME ZONE 'UTC')::int AS weekday,
		       EXTRACT(HOUR FROM clicked_at AT TIME ZONE 'UTC')::int AS hour,
		       COUNT(*)
		FROM email_logs
		WHERE project_id = $1 AND clicked_at IS NOT NULL AND clicked_at >= $2 AND clicked_at < $3` + nlFilter + `
		GROUP BY weekday, hour
		ORDER BY weekday, hour
	`, append([]any{projectID, from, to}, nlArgs...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cells := []HeatCell{}
	for rows.Next() {
		var c HeatCell
		if err := rows.Scan(&c.Weekday, &c.Hour, &c.Count); err != nil {
			return nil, err
		}
		cells = append(cells, c)
	}
	return cells, rows.Err()
}

func sortedBreakdown(counts map[string]int64) []Breakdown {
	out := make([]Breakdown, 0, len(counts))
	for label, count := range counts {
		out = append(out, Breakdown{Label: label, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Label < out[j].Label
	})
	return out
}

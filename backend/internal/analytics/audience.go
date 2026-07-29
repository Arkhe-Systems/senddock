package analytics

import (
	"fmt"
	"net/http"
	"sort"
	"time"
)

type AudienceBucket struct {
	Bucket        string `json:"bucket"`
	Added         int64  `json:"added"`
	Unsubscribed  int64  `json:"unsubscribed"`
	NetGrowth     int64  `json:"net_growth"`
	CumulativeNet int64  `json:"cumulative_net"`
}

type Audience struct {
	From              time.Time        `json:"from"`
	To                time.Time        `json:"to"`
	Granularity       string           `json:"granularity"`
	ActiveTotal       int64            `json:"active_total"`
	UnsubscribedTotal int64            `json:"unsubscribed_total"`
	AddedInRange      int64            `json:"added_in_range"`
	UnsubInRange      int64            `json:"unsubscribed_in_range"`
	Series            []AudienceBucket `json:"series"`
}

func (h *Handler) Audience(w http.ResponseWriter, r *http.Request) {
	projectID, ok := h.authorizedProjectFromRequest(w, r)
	if !ok {
		return
	}
	from, to, granularity, ok := parseWindow(w, r)
	if !ok {
		return
	}

	out := Audience{From: from, To: to, Granularity: granularity, Series: []AudienceBucket{}}

	if err := h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM subscribers WHERE project_id = $1 AND status = 'active'`,
		projectID).Scan(&out.ActiveTotal); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load audience")
		return
	}
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM subscribers WHERE project_id = $1 AND status = 'unsubscribed'`,
		projectID).Scan(&out.UnsubscribedTotal); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load audience")
		return
	}

	added, err := h.bucketCounts(r, `
		SELECT to_char(date_trunc('%s', created_at) AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS bucket, COUNT(*)
		FROM subscribers
		WHERE project_id = $1 AND created_at >= $2 AND created_at < $3
		GROUP BY bucket`, projectID, from, to, granularity)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load sign-ups")
		return
	}
	unsub, err := h.bucketCounts(r, `
		SELECT to_char(date_trunc('%s', unsubscribed_at) AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS bucket, COUNT(*)
		FROM subscribers
		WHERE project_id = $1 AND unsubscribed_at IS NOT NULL AND unsubscribed_at >= $2 AND unsubscribed_at < $3
		GROUP BY bucket`, projectID, from, to, granularity)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load unsubscribes")
		return
	}

	buckets := map[string]bool{}
	for b := range added {
		buckets[b] = true
	}
	for b := range unsub {
		buckets[b] = true
	}
	ordered := make([]string, 0, len(buckets))
	for b := range buckets {
		ordered = append(ordered, b)
	}
	sort.Strings(ordered)

	var cumulative int64
	for _, b := range ordered {
		net := added[b] - unsub[b]
		cumulative += net
		out.AddedInRange += added[b]
		out.UnsubInRange += unsub[b]
		out.Series = append(out.Series, AudienceBucket{
			Bucket:        b,
			Added:         added[b],
			Unsubscribed:  unsub[b],
			NetGrowth:     net,
			CumulativeNet: cumulative,
		})
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) bucketCounts(r *http.Request, query, projectID string, from, to time.Time, granularity string) (map[string]int64, error) {
	rows, err := h.db.QueryContext(r.Context(), fmt.Sprintf(query, granularity), projectID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var bucket string
		var count int64
		if err := rows.Scan(&bucket, &count); err != nil {
			return nil, err
		}
		out[bucket] = count
	}
	return out, rows.Err()
}

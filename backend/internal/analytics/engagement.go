package analytics

import (
	"net/http"
	"sort"
)

type Breakdown struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type Engagement struct {
	Devices []Breakdown `json:"devices"`
	Clients []Breakdown `json:"clients"`
}

// Engagement breaks recorded clicks down by device and mail client, parsed
// from email_clicks.user_agent — data that is captured on every click but was
// never surfaced. Grouping happens in Go because the classification is a
// heuristic, not something to encode in SQL.
func (h *Handler) Engagement(w http.ResponseWriter, r *http.Request) {
	projectID, ok := h.authorizedProjectFromRequest(w, r)
	if !ok {
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT COALESCE(ec.user_agent, '')
		FROM email_clicks ec
		JOIN email_logs l ON l.id = ec.log_id
		WHERE l.project_id = $1
	`, projectID)
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

	writeJSON(w, http.StatusOK, Engagement{
		Devices: sortedBreakdown(devices),
		Clients: sortedBreakdown(clients),
	})
}

// sortedBreakdown turns a count map into a slice ordered by count desc, then
// label, so the output is stable.
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

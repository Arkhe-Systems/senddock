package analytics

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	projectID, ok := h.authorizedProjectFromRequest(w, r)
	if !ok {
		return
	}

	rows, err := h.db.QueryContext(r.Context(), campaignSelect+`
		WHERE b.project_id = $1
		GROUP BY b.id
		ORDER BY b.started_at DESC
		LIMIT 5000
	`, projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to export campaigns")
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="senddock-campaigns-`+time.Now().UTC().Format("2006-01-02")+`.csv"`)

	cw := csv.NewWriter(w)
	defer cw.Flush()
	_ = cw.Write([]string{
		"broadcast_id", "subject", "status", "started_at", "finished_at",
		"total_recipients", "sent", "failed", "bounced", "opened", "clicked",
		"acceptance_pct", "bounce_rate_pct", "open_rate_pct", "click_rate_pct", "click_to_open_pct",
	})

	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return
		}
		_ = cw.Write([]string{
			c.BroadcastID, c.Subject, c.Status, c.StartedAt, c.FinishedAt,
			strconv.FormatInt(c.TotalRecipients, 10),
			strconv.FormatInt(c.Sent, 10), strconv.FormatInt(c.Failed, 10), strconv.FormatInt(c.Bounced, 10),
			strconv.FormatInt(c.Opened, 10), strconv.FormatInt(c.Clicked, 10),
			pct(c.AcceptancePct), pct(c.BounceRatePct), pct(c.OpenRatePct), pct(c.ClickRatePct), pct(c.ClickToOpenPct),
		})
	}
}

func pct(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

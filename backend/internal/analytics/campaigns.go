package analytics

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/google/uuid"
)

type CampaignStat struct {
	BroadcastID     string  `json:"broadcast_id"`
	Subject         string  `json:"subject"`
	Status          string  `json:"status"`
	StartedAt       string  `json:"started_at"`
	FinishedAt      string  `json:"finished_at,omitempty"`
	TotalRecipients int64   `json:"total_recipients"`
	Sent            int64   `json:"sent"`
	Failed          int64   `json:"failed"`
	Bounced         int64   `json:"bounced"`
	Opened          int64   `json:"opened"`
	Clicked         int64   `json:"clicked"`
	AcceptancePct   float64 `json:"acceptance_pct"`
	BounceRatePct   float64 `json:"bounce_rate_pct"`
	OpenRatePct     float64 `json:"open_rate_pct"`
	ClickRatePct    float64 `json:"click_rate_pct"`
	ClickToOpenPct  float64 `json:"click_to_open_pct"`
}

type CampaignDetail struct {
	CampaignStat
	TopClickedLinks []LinkStat `json:"top_clicked_links"`
}

const campaignSelect = `
	SELECT
		b.id, b.subject, b.status, b.started_at, b.finished_at, b.total_recipients,
		COUNT(*) FILTER (WHERE l.status = 'sent')        AS sent,
		COUNT(*) FILTER (WHERE l.status = 'failed')      AS failed,
		COUNT(*) FILTER (WHERE l.status = 'bounced')     AS bounced,
		COUNT(*) FILTER (WHERE l.opened_at IS NOT NULL)  AS opened,
		COUNT(*) FILTER (WHERE l.clicked_at IS NOT NULL) AS clicked
	FROM broadcasts b
	LEFT JOIN email_logs l ON l.broadcast_id = b.id
`

func (h *Handler) Campaigns(w http.ResponseWriter, r *http.Request) {
	projectID, ok := h.authorizedProjectFromRequest(w, r)
	if !ok {
		return
	}
	from, to, _, ok := parseWindow(w, r)
	if !ok {
		return
	}

	rows, err := h.db.QueryContext(r.Context(), campaignSelect+`
		WHERE b.project_id = $1 AND b.started_at >= $2 AND b.started_at < $3
		GROUP BY b.id
		ORDER BY b.started_at DESC
		LIMIT 200
	`, projectID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load campaigns")
		return
	}
	defer rows.Close()

	campaigns := []CampaignStat{}
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read campaigns")
			return
		}
		campaigns = append(campaigns, c)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read campaigns")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"campaigns": campaigns})
}

func (h *Handler) Campaign(w http.ResponseWriter, r *http.Request) {
	projectID, ok := h.authorizedProjectFromRequest(w, r)
	if !ok {
		return
	}

	broadcastID := r.PathValue("broadcastId")
	if _, err := uuid.Parse(broadcastID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid broadcast id")
		return
	}

	row := h.db.QueryRowContext(r.Context(), campaignSelect+`
		WHERE b.project_id = $1 AND b.id = $2
		GROUP BY b.id
	`, projectID, broadcastID)

	c, err := scanCampaign(row)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load campaign")
		return
	}

	detail := CampaignDetail{CampaignStat: c, TopClickedLinks: []LinkStat{}}

	linkRows, err := h.db.QueryContext(r.Context(), `
		SELECT ec.url, COUNT(*) AS clicks
		FROM email_clicks ec
		JOIN email_logs l ON l.id = ec.log_id
		WHERE l.broadcast_id = $1
		GROUP BY ec.url
		ORDER BY clicks DESC
		LIMIT 20
	`, broadcastID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load campaign links")
		return
	}
	defer linkRows.Close()
	for linkRows.Next() {
		var ls LinkStat
		if err := linkRows.Scan(&ls.URL, &ls.Clicks); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read campaign links")
			return
		}
		detail.TopClickedLinks = append(detail.TopClickedLinks, ls)
	}

	writeJSON(w, http.StatusOK, detail)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCampaign(s rowScanner) (CampaignStat, error) {
	var c CampaignStat
	var startedAt time.Time
	var finishedAt sql.NullTime
	if err := s.Scan(
		&c.BroadcastID, &c.Subject, &c.Status, &startedAt, &finishedAt, &c.TotalRecipients,
		&c.Sent, &c.Failed, &c.Bounced, &c.Opened, &c.Clicked,
	); err != nil {
		return c, err
	}
	c.StartedAt = startedAt.UTC().Format(time.RFC3339)
	if finishedAt.Valid {
		c.FinishedAt = finishedAt.Time.UTC().Format(time.RFC3339)
	}
	c.AcceptancePct = acceptancePct(c.Sent, c.Failed, c.Bounced)
	c.BounceRatePct = bounceRatePct(c.Sent, c.Failed, c.Bounced)
	c.OpenRatePct = ratePct(c.Opened, c.Sent)
	c.ClickRatePct = ratePct(c.Clicked, c.Sent)
	c.ClickToOpenPct = ratePct(c.Clicked, c.Opened)
	return c, nil
}

func (h *Handler) authorizedProjectFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return "", false
	}
	projectID := r.PathValue("id")
	if err := h.authorizeProject(r.Context(), projectID, userID); err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			writeError(w, http.StatusInternalServerError, "authorization failed")
		}
		return "", false
	}
	return projectID, true
}

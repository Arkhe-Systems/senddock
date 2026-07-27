package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/arkhe-systems/senddock/pkg/segments"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

type Overview struct {
	From               time.Time           `json:"from"`
	To                 time.Time           `json:"to"`
	Granularity        string              `json:"granularity"`
	RangeDays          int                 `json:"range_days"`
	SegmentID          string              `json:"segment_id,omitempty"`
	TotalSent          int64               `json:"total_sent"`
	TotalFailed        int64               `json:"total_failed"`
	TotalBounced       int64               `json:"total_bounced"`
	TotalOpened        int64               `json:"total_opened"`
	TotalClicked       int64               `json:"total_clicked"`
	AcceptancePct  float64             `json:"acceptance_pct"`
	BounceRatePct      float64             `json:"bounce_rate_pct"`
	OpenRatePct        float64             `json:"open_rate_pct"`
	ClickRatePct       float64             `json:"click_rate_pct"`
	ClickToOpenPct     float64             `json:"click_to_open_pct"`
	OpensSeries        []OpenBucket        `json:"opens_series"`
	ClicksSeries       []ClickBucket       `json:"clicks_series"`
	TopTemplates       []TemplateStat      `json:"top_templates"`
	TopClickedLinks    []LinkStat          `json:"top_clicked_links"`
	ActiveSubscribers  int64               `json:"active_subscribers"`
	SendsByStatus      []StatusBucket      `json:"sends_by_status"`
	Previous           PeriodMetrics       `json:"previous"`
	BroadcastsInFlight []BroadcastInFlight `json:"broadcasts_in_flight"`
}

type BroadcastInFlight struct {
	ID         string    `json:"id"`
	Subject    string    `json:"subject"`
	Total      int       `json:"total"`
	Sent       int       `json:"sent"`
	Failed     int       `json:"failed"`
	Suppressed int       `json:"suppressed"`
	Remaining  int       `json:"remaining"`
	StartedAt  time.Time `json:"started_at"`
}

type PeriodMetrics struct {
	TotalSent         int64   `json:"total_sent"`
	TotalFailed       int64   `json:"total_failed"`
	TotalBounced      int64   `json:"total_bounced"`
	TotalOpened       int64   `json:"total_opened"`
	TotalClicked      int64   `json:"total_clicked"`
	AcceptancePct float64 `json:"acceptance_pct"`
	BounceRatePct     float64 `json:"bounce_rate_pct"`
	OpenRatePct       float64 `json:"open_rate_pct"`
	ClickRatePct      float64 `json:"click_rate_pct"`
}

type LinkStat struct {
	URL    string `json:"url"`
	Clicks int64  `json:"clicks"`
}

type OpenBucket struct {
	Bucket string `json:"bucket"`
	Opens  int64  `json:"opens"`
}

type ClickBucket struct {
	Bucket string `json:"bucket"`
	Clicks int64  `json:"clicks"`
}

type TemplateStat struct {
	TemplateID string `json:"template_id"`
	Name       string `json:"name"`
	Sends      int64  `json:"sends"`
}

type StatusBucket struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

var ErrProjectNotFound = errors.New("project not found")
var ErrForbidden = errors.New("forbidden")

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
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
		return
	}

	from, to, granularity, ok := parseWindow(w, r)
	if !ok {
		return
	}

	var subIDs []uuid.UUID
	segmentID := strings.TrimSpace(r.URL.Query().Get("segment_id"))
	if segmentID != "" {
		ids, err := h.segmentSubscriberIDs(r.Context(), projectID, segmentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "segment not found")
				return
			}
			if errors.Is(err, segments.ErrInvalidPredicate) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, "invalid segment_id")
			return
		}
		subIDs = ids
	}

	overview, err := h.computeOverview(r.Context(), projectID, from, to, granularity, subIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute analytics")
		return
	}
	overview.SegmentID = segmentID

	writeJSON(w, http.StatusOK, overview)
}

func (h *Handler) segmentSubscriberIDs(ctx context.Context, projectID, segmentID string) ([]uuid.UUID, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	sid, err := uuid.Parse(segmentID)
	if err != nil {
		return nil, errors.New("invalid segment id")
	}

	var raw json.RawMessage
	if err := h.db.QueryRowContext(ctx, `
		SELECT predicate FROM segments WHERE id = $1 AND project_id = $2
	`, sid, pid).Scan(&raw); err != nil {
		return nil, err
	}

	pred, err := segments.ParsePredicate(raw)
	if err != nil {
		return nil, err
	}

	query := "SELECT id FROM subscribers WHERE project_id = $1"
	args := []any{pid}
	where, whereArgs := segments.BuildWhere(pred, len(args)+1)
	if where != "" {
		query += " AND (" + where + ")"
		args = append(args, whereArgs...)
	}

	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func autoGranularity(from, to time.Time) string {
	d := to.Sub(from)
	switch {
	case d <= 48*time.Hour:
		return "hour"
	case d <= 90*24*time.Hour:
		return "day"
	case d <= 365*24*time.Hour:
		return "week"
	default:
		return "month"
	}
}

func (h *Handler) authorizeProject(ctx context.Context, projectID, userID string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return ErrProjectNotFound
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrForbidden
	}

	var found uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		SELECT p.id
		FROM projects p
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&found)
	if err == sql.ErrNoRows {
		return ErrForbidden
	}
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) computeOverview(ctx context.Context, projectID string, from, to time.Time, granularity string, subIDs []uuid.UUID) (Overview, error) {
	since := from
	until := to
	days := int(to.Sub(from).Hours()/24) + 1

	out := Overview{
		From:               from,
		To:                 to,
		Granularity:        granularity,
		RangeDays:          days,
		OpensSeries:        []OpenBucket{},
		ClicksSeries:       []ClickBucket{},
		TopTemplates:       []TemplateStat{},
		TopClickedLinks:    []LinkStat{},
		SendsByStatus:      []StatusBucket{},
		BroadcastsInFlight: []BroadcastInFlight{},
	}

	logFilter := ""
	logArgs := []any{projectID, since, until}
	if subIDs != nil {
		logFilter = " AND subscriber_id = ANY($4)"
		logArgs = append(logArgs, pq.Array(subIDs))
	}

	statusRows, err := h.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM email_logs
		WHERE project_id = $1 AND sent_at >= $2 AND sent_at < $3`+logFilter+`
		GROUP BY status
	`, logArgs...)
	if err != nil {
		return out, err
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var b StatusBucket
		if err := statusRows.Scan(&b.Status, &b.Count); err != nil {
			return out, err
		}
		out.SendsByStatus = append(out.SendsByStatus, b)
		switch b.Status {
		case "sent":
			out.TotalSent = b.Count
		case "failed":
			out.TotalFailed = b.Count
		case "bounced":
			out.TotalBounced = b.Count
		}
	}
	if err := statusRows.Err(); err != nil {
		return out, err
	}

	out.AcceptancePct = acceptancePct(out.TotalSent, out.TotalFailed, out.TotalBounced)
	out.BounceRatePct = bounceRatePct(out.TotalSent, out.TotalFailed, out.TotalBounced)

	if err := h.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM email_logs
		WHERE project_id = $1 AND sent_at >= $2 AND sent_at < $3 AND opened_at IS NOT NULL`+logFilter+`
	`, logArgs...).Scan(&out.TotalOpened); err != nil {
		return out, err
	}
	out.OpenRatePct = ratePct(out.TotalOpened, out.TotalSent)

	openRows, err := h.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT to_char(date_trunc('%s', opened_at) AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS bucket, COUNT(*)
		FROM email_logs
		WHERE project_id = $1 AND opened_at IS NOT NULL AND opened_at >= $2 AND opened_at < $3%s
		GROUP BY bucket
		ORDER BY bucket
	`, granularity, logFilter), logArgs...)
	if err != nil {
		return out, err
	}
	defer openRows.Close()

	for openRows.Next() {
		var b OpenBucket
		if err := openRows.Scan(&b.Bucket, &b.Opens); err != nil {
			return out, err
		}
		out.OpensSeries = append(out.OpensSeries, b)
	}
	if err := openRows.Err(); err != nil {
		return out, err
	}

	clickRows, err := h.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT to_char(date_trunc('%s', clicked_at) AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS bucket, COUNT(*)
		FROM email_logs
		WHERE project_id = $1 AND clicked_at IS NOT NULL AND clicked_at >= $2 AND clicked_at < $3%s
		GROUP BY bucket
		ORDER BY bucket
	`, granularity, logFilter), logArgs...)
	if err != nil {
		return out, err
	}
	defer clickRows.Close()

	for clickRows.Next() {
		var b ClickBucket
		if err := clickRows.Scan(&b.Bucket, &b.Clicks); err != nil {
			return out, err
		}
		out.ClicksSeries = append(out.ClicksSeries, b)
	}
	if err := clickRows.Err(); err != nil {
		return out, err
	}

	tplFilter := ""
	if subIDs != nil {
		tplFilter = " AND l.subscriber_id = ANY($4)"
	}
	tplRows, err := h.db.QueryContext(ctx, `
		SELECT t.id, t.name, COUNT(l.*) AS sends
		FROM email_logs l
		JOIN templates t ON t.id = l.template_id
		WHERE l.project_id = $1 AND l.sent_at >= $2 AND l.sent_at < $3`+tplFilter+`
		GROUP BY t.id, t.name
		ORDER BY sends DESC
		LIMIT 10
	`, logArgs...)
	if err != nil {
		return out, err
	}
	defer tplRows.Close()

	for tplRows.Next() {
		var s TemplateStat
		if err := tplRows.Scan(&s.TemplateID, &s.Name, &s.Sends); err != nil {
			return out, err
		}
		out.TopTemplates = append(out.TopTemplates, s)
	}
	if err := tplRows.Err(); err != nil {
		return out, err
	}

	subsQuery := `
		SELECT COUNT(*) FROM subscribers
		WHERE project_id = $1 AND status = 'active'`
	subsArgs := []any{projectID}
	if subIDs != nil {
		subsQuery += " AND id = ANY($2)"
		subsArgs = append(subsArgs, pq.Array(subIDs))
	}
	if err := h.db.QueryRowContext(ctx, subsQuery, subsArgs...).Scan(&out.ActiveSubscribers); err != nil {
		return out, err
	}

	if err := h.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM email_logs
		WHERE project_id = $1 AND sent_at >= $2 AND sent_at < $3 AND clicked_at IS NOT NULL`+logFilter+`
	`, logArgs...).Scan(&out.TotalClicked); err != nil {
		return out, err
	}
	out.ClickRatePct = ratePct(out.TotalClicked, out.TotalSent)
	out.ClickToOpenPct = ratePct(out.TotalClicked, out.TotalOpened)

	linkFilter := ""
	if subIDs != nil {
		linkFilter = " AND l.subscriber_id = ANY($4)"
	}
	linkRows, err := h.db.QueryContext(ctx, `
		SELECT c.url, COUNT(*) AS clicks
		FROM email_clicks c
		JOIN email_logs l ON l.id = c.log_id
		WHERE l.project_id = $1 AND c.clicked_at >= $2 AND c.clicked_at < $3`+linkFilter+`
		GROUP BY c.url
		ORDER BY clicks DESC
		LIMIT 10
	`, logArgs...)
	if err != nil {
		return out, err
	}
	defer linkRows.Close()

	for linkRows.Next() {
		var s LinkStat
		if err := linkRows.Scan(&s.URL, &s.Clicks); err != nil {
			return out, err
		}
		out.TopClickedLinks = append(out.TopClickedLinks, s)
	}
	if err := linkRows.Err(); err != nil {
		return out, err
	}

	flightRows, err := h.db.QueryContext(ctx, `
		SELECT id, subject, total_recipients, sent_count, failed_count, suppressed_count, started_at
		FROM broadcasts
		WHERE project_id = $1 AND status = 'sending'
		ORDER BY started_at DESC
		LIMIT 20
	`, projectID)
	if err != nil {
		return out, err
	}
	defer flightRows.Close()

	for flightRows.Next() {
		var b BroadcastInFlight
		if err := flightRows.Scan(&b.ID, &b.Subject, &b.Total, &b.Sent, &b.Failed, &b.Suppressed, &b.StartedAt); err != nil {
			return out, err
		}
		b.Remaining = b.Total - (b.Sent + b.Failed + b.Suppressed)
		if b.Remaining < 0 {
			b.Remaining = 0
		}
		out.BroadcastsInFlight = append(out.BroadcastsInFlight, b)
	}
	if err := flightRows.Err(); err != nil {
		return out, err
	}

	duration := until.Sub(since)
	prev, err := h.periodMetrics(ctx, projectID, since.Add(-duration), since, subIDs)
	if err != nil {
		return out, err
	}
	out.Previous = prev

	return out, nil
}

func (h *Handler) periodMetrics(ctx context.Context, projectID string, since, until time.Time, subIDs []uuid.UUID) (PeriodMetrics, error) {
	var m PeriodMetrics

	logFilter := ""
	logArgs := []any{projectID, since, until}
	if subIDs != nil {
		logFilter = " AND subscriber_id = ANY($4)"
		logArgs = append(logArgs, pq.Array(subIDs))
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM email_logs
		WHERE project_id = $1 AND sent_at >= $2 AND sent_at < $3`+logFilter+`
		GROUP BY status
	`, logArgs...)
	if err != nil {
		return m, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return m, err
		}
		switch status {
		case "sent":
			m.TotalSent = count
		case "failed":
			m.TotalFailed = count
		case "bounced":
			m.TotalBounced = count
		}
	}
	if err := rows.Err(); err != nil {
		return m, err
	}

	m.AcceptancePct = acceptancePct(m.TotalSent, m.TotalFailed, m.TotalBounced)
	m.BounceRatePct = bounceRatePct(m.TotalSent, m.TotalFailed, m.TotalBounced)

	if err := h.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM email_logs
		WHERE project_id = $1 AND sent_at >= $2 AND sent_at < $3 AND opened_at IS NOT NULL`+logFilter+`
	`, logArgs...).Scan(&m.TotalOpened); err != nil {
		return m, err
	}
	m.OpenRatePct = ratePct(m.TotalOpened, m.TotalSent)

	if err := h.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM email_logs
		WHERE project_id = $1 AND sent_at >= $2 AND sent_at < $3 AND clicked_at IS NOT NULL`+logFilter+`
	`, logArgs...).Scan(&m.TotalClicked); err != nil {
		return m, err
	}
	m.ClickRatePct = ratePct(m.TotalClicked, m.TotalSent)

	return m, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

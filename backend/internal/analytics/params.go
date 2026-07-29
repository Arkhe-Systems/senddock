package analytics

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseWindow(w http.ResponseWriter, r *http.Request) (from, to time.Time, granularity string, ok bool) {
	now := time.Now().UTC()
	to = now
	from = now.AddDate(0, 0, -30)

	if v := r.URL.Query().Get("range_days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 3650 {
			from = to.AddDate(0, 0, -n)
		}
	}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t.UTC()
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t.UTC()
		}
	}
	if !to.After(from) {
		writeError(w, http.StatusBadRequest, "to must be after from")
		return time.Time{}, time.Time{}, "", false
	}

	granularity = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("granularity")))
	switch granularity {
	case "hour", "day", "week", "month":
	case "", "auto":
		granularity = autoGranularity(from, to)
	default:
		writeError(w, http.StatusBadRequest, "granularity must be one of: hour, day, week, month, auto")
		return time.Time{}, time.Time{}, "", false
	}

	return from, to, granularity, true
}

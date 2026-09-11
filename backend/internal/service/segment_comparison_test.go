package service

import (
	"context"
	"strconv"
	"testing"

	"github.com/arkhe-systems/senddock/pkg/segments"
)

// The comparison a date rule compiles to has to run against a real database, because that
// is where the old cast failed: the segment saved fine and then took the preview and the
// broadcast down with an invalid-input-syntax error.
func TestSegmentDateComparisonRunsAgainstPostgres(t *testing.T) {
	conn, _ := leaseTestDB(t)
	ctx := context.Background()

	_, projectID, _ := leaseFixtures(t, conn)

	subscribers := []struct{ email, signup string }{
		{"before@example.test", "2025-12-31"},
		{"after@example.test", "2026-02-01"},
	}
	for _, s := range subscribers {
		if _, err := conn.ExecContext(ctx,
			`INSERT INTO subscribers (project_id, email, metadata) VALUES ($1, $2, $3::jsonb)`,
			projectID, s.email, `{"signup":"`+s.signup+`"}`,
		); err != nil {
			t.Fatalf("insert subscriber %s: %v", s.email, err)
		}
	}

	where, args := segments.BuildWhere(segments.Predicate{
		Match: "all",
		Rules: []segments.Rule{{Field: "custom.signup", Op: "gt", Value: "2026-01-01"}},
	}, 2)

	var matched int
	query := `SELECT COUNT(*) FROM subscribers WHERE project_id = $1 AND ` + where
	if err := conn.QueryRowContext(ctx, query, append([]any{projectID}, args...)...).Scan(&matched); err != nil {
		t.Fatalf("date comparison failed to run: %v\nquery: %s", err, query)
	}
	if matched != 1 {
		t.Fatalf("date comparison matched %d subscribers, want 1 (only the 2026-02-01 signup)", matched)
	}
}

// The numeric path has to keep working alongside it.
func TestSegmentNumberComparisonRunsAgainstPostgres(t *testing.T) {
	conn, _ := leaseTestDB(t)
	ctx := context.Background()

	_, projectID, _ := leaseFixtures(t, conn)

	for _, s := range []struct {
		email string
		seats int
	}{{"small@example.test", 3}, {"big@example.test", 50}} {
		if _, err := conn.ExecContext(ctx,
			`INSERT INTO subscribers (project_id, email, metadata) VALUES ($1, $2, $3::jsonb)`,
			projectID, s.email, `{"seats":`+strconv.Itoa(s.seats)+`}`,
		); err != nil {
			t.Fatalf("insert subscriber %s: %v", s.email, err)
		}
	}

	where, args := segments.BuildWhere(segments.Predicate{
		Match: "all",
		Rules: []segments.Rule{{Field: "custom.seats", Op: "gt", Value: float64(10)}},
	}, 2)

	var matched int
	query := `SELECT COUNT(*) FROM subscribers WHERE project_id = $1 AND ` + where
	if err := conn.QueryRowContext(ctx, query, append([]any{projectID}, args...)...).Scan(&matched); err != nil {
		t.Fatalf("number comparison failed to run: %v\nquery: %s", err, query)
	}
	if matched != 1 {
		t.Fatalf("number comparison matched %d subscribers, want 1", matched)
	}
}

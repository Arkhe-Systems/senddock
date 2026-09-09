package service

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/arkhe-systems/senddock/internal/db"
)

// These tests exercise the broadcast job ownership lease against a real database.
// They skip unless DATABASE_URL points at a migrated instance, so the default test run
// stays free of external dependencies; CI passes the value explicitly when a database
// is available.

func leaseTestDB(t *testing.T) (*sql.DB, *db.Queries) {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping broadcast lease integration test")
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		t.Fatalf("ping database: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return conn, db.New(conn)
}

// leaseFixtures creates the minimum graph a broadcast job needs and returns the ids.
// Deleting the workspace in cleanup cascades to everything the test inserted.
func leaseFixtures(t *testing.T, conn *sql.DB) (broadcastID, projectID, subscriberID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	var userID, workspaceID, templateID uuid.UUID

	if err := conn.QueryRowContext(ctx,
		`INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id`,
		"lease-test-"+uuid.NewString()+"@example.test", "Lease Test",
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if err := conn.QueryRowContext(ctx,
		`INSERT INTO workspaces (name, created_by) VALUES ($1, $2) RETURNING id`,
		"lease-test", userID,
	).Scan(&workspaceID); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	t.Cleanup(func() {
		if _, err := conn.ExecContext(ctx, `DELETE FROM workspaces WHERE id = $1`, workspaceID); err != nil {
			t.Logf("cleanup workspace: %v", err)
		}
	})

	if err := conn.QueryRowContext(ctx,
		`INSERT INTO projects (workspace_id, user_id, name, from_email) VALUES ($1, $2, $3, $4) RETURNING id`,
		workspaceID, userID, "lease-test", "sender@example.test",
	).Scan(&projectID); err != nil {
		t.Fatalf("insert project: %v", err)
	}

	if err := conn.QueryRowContext(ctx,
		`INSERT INTO templates (project_id, name, subject) VALUES ($1, $2, $3) RETURNING id`,
		projectID, "lease-test", "Lease test",
	).Scan(&templateID); err != nil {
		t.Fatalf("insert template: %v", err)
	}

	if err := conn.QueryRowContext(ctx,
		`INSERT INTO subscribers (project_id, email) VALUES ($1, $2) RETURNING id`,
		projectID, "recipient@example.test",
	).Scan(&subscriberID); err != nil {
		t.Fatalf("insert subscriber: %v", err)
	}

	if err := conn.QueryRowContext(ctx,
		`INSERT INTO broadcasts (project_id, template_id) VALUES ($1, $2) RETURNING id`,
		projectID, templateID,
	).Scan(&broadcastID); err != nil {
		t.Fatalf("insert broadcast: %v", err)
	}

	return broadcastID, projectID, subscriberID
}

// insertSendingJob writes a job already in 'sending' with the given ownership, which is
// what startup recovery inspects.
func insertSendingJob(t *testing.T, conn *sql.DB, broadcastID, projectID, subscriberID uuid.UUID, recipient string, worker uuid.NullUUID, lease sql.NullTime) uuid.UUID {
	t.Helper()

	var id uuid.UUID
	if err := conn.QueryRowContext(context.Background(),
		`INSERT INTO broadcast_jobs
		     (broadcast_id, project_id, subscriber_id, recipient_email, status, worker_id, lease_expires_at)
		 VALUES ($1, $2, $3, $4, 'sending', $5, $6)
		 RETURNING id`,
		broadcastID, projectID, subscriberID, recipient, worker, lease,
	).Scan(&id); err != nil {
		t.Fatalf("insert sending job: %v", err)
	}
	return id
}

func jobOwnerAndLease(t *testing.T, conn *sql.DB, id uuid.UUID) (string, uuid.NullUUID, sql.NullTime) {
	t.Helper()

	var status string
	var worker uuid.NullUUID
	var lease sql.NullTime
	if err := conn.QueryRowContext(context.Background(),
		`SELECT status, worker_id, lease_expires_at FROM broadcast_jobs WHERE id = $1`, id,
	).Scan(&status, &worker, &lease); err != nil {
		t.Fatalf("read job: %v", err)
	}
	return status, worker, lease
}

// A restart must not reclaim a job whose lease is still valid: that job belongs to an
// instance that is actively delivering it, and reclaiming it sends the email twice.
func TestResetStuckSendingJobsLeavesLiveLeasesAlone(t *testing.T) {
	conn, q := leaseTestDB(t)
	broadcastID, projectID, subscriberID := leaseFixtures(t, conn)
	ctx := context.Background()

	liveWorker := uuid.New()
	liveID := insertSendingJob(t, conn, broadcastID, projectID, subscriberID,
		"live@example.test",
		uuid.NullUUID{UUID: liveWorker, Valid: true},
		sql.NullTime{Time: time.Now().Add(time.Minute), Valid: true})

	// Recover as if this were a fresh process starting up.
	if _, err := q.ResetStuckSendingJobs(ctx, time.Now()); err != nil {
		t.Fatalf("reset stuck jobs: %v", err)
	}

	status, worker, lease := jobOwnerAndLease(t, conn, liveID)
	if status != "sending" {
		t.Fatalf("job with a live lease was reclaimed: status = %q, want %q", status, "sending")
	}
	if !worker.Valid || worker.UUID != liveWorker {
		t.Fatalf("job with a live lease lost its owner: worker = %+v, want %s", worker, liveWorker)
	}
	if !lease.Valid {
		t.Fatal("job with a live lease lost its deadline")
	}
}

// Abandoned work still has to be recovered: an expired lease and a row written before
// leases existed both go back to the queue.
func TestResetStuckSendingJobsReclaimsAbandonedWork(t *testing.T) {
	conn, q := leaseTestDB(t)
	broadcastID, projectID, subscriberID := leaseFixtures(t, conn)
	ctx := context.Background()

	expiredID := insertSendingJob(t, conn, broadcastID, projectID, subscriberID,
		"expired@example.test",
		uuid.NullUUID{UUID: uuid.New(), Valid: true},
		sql.NullTime{Time: time.Now().Add(-time.Minute), Valid: true})

	legacyID := insertSendingJob(t, conn, broadcastID, projectID, subscriberID,
		"legacy@example.test",
		uuid.NullUUID{}, sql.NullTime{})

	reclaimed, err := q.ResetStuckSendingJobs(ctx, time.Now())
	if err != nil {
		t.Fatalf("reset stuck jobs: %v", err)
	}
	if reclaimed != 2 {
		t.Fatalf("reclaimed %d jobs, want 2 (expired lease + pre-lease row)", reclaimed)
	}

	for name, id := range map[string]uuid.UUID{"expired": expiredID, "legacy": legacyID} {
		status, worker, lease := jobOwnerAndLease(t, conn, id)
		if status != "retry" {
			t.Errorf("%s job: status = %q, want %q", name, status, "retry")
		}
		if worker.Valid || lease.Valid {
			t.Errorf("%s job: ownership not cleared (worker=%+v lease=%+v)", name, worker, lease)
		}
	}
}

// A claimed job carries its owner and deadline, which is what makes the recovery above
// able to tell abandoned work from work in flight.
func TestClaimBroadcastJobStampsOwnerAndLease(t *testing.T) {
	conn, q := leaseTestDB(t)
	broadcastID, projectID, subscriberID := leaseFixtures(t, conn)
	ctx := context.Background()

	jobID := insertSendingJob(t, conn, broadcastID, projectID, subscriberID,
		"pending@example.test", uuid.NullUUID{}, sql.NullTime{})
	if _, err := conn.ExecContext(ctx,
		`UPDATE broadcast_jobs SET status = 'pending', scheduled_at = NOW() - INTERVAL '1 minute' WHERE id = $1`,
		jobID); err != nil {
		t.Fatalf("make job claimable: %v", err)
	}

	worker := uuid.New()
	deadline := time.Now().Add(broadcastJobLease)
	claimed, err := q.ClaimBroadcastJob(ctx, db.ClaimBroadcastJobParams{
		WorkerID:       worker,
		LeaseExpiresAt: deadline,
	})
	if err != nil {
		t.Fatalf("claim job: %v", err)
	}
	if claimed.ID != jobID {
		t.Fatalf("claimed job %s, want %s", claimed.ID, jobID)
	}
	if !claimed.WorkerID.Valid || claimed.WorkerID.UUID != worker {
		t.Fatalf("claimed job owner = %+v, want %s", claimed.WorkerID, worker)
	}
	if !claimed.LeaseExpiresAt.Valid {
		t.Fatal("claimed job carries no deadline")
	}
	if diff := claimed.LeaseExpiresAt.Time.Sub(deadline).Abs(); diff > time.Second {
		t.Fatalf("claimed job deadline = %s, want ~%s (off by %s)", claimed.LeaseExpiresAt.Time, deadline, diff)
	}
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

const (
	broadcastWorkerCount     = 5
	broadcastWorkerIdleSleep = 2 * time.Second
	broadcastJobMaxAttempts  = 5
)

type BroadcastWorker struct {
	queries      *db.Queries
	emailService *EmailService
}

func NewBroadcastWorker(queries *db.Queries, emailService *EmailService) *BroadcastWorker {
	return &BroadcastWorker{queries: queries, emailService: emailService}
}

func (w *BroadcastWorker) Start(ctx context.Context) {
	if rows, err := w.queries.ResetStuckSendingJobs(ctx); err != nil {
		slog.Error("broadcast worker: reset stuck jobs failed", "error", err)
	} else if rows > 0 {
		slog.Warn("broadcast worker: reset stuck sending jobs to retry", "jobs", rows)
	}

	for i := 0; i < broadcastWorkerCount; i++ {
		go w.run(ctx, i)
	}
	slog.Info("broadcast worker started", "goroutines", broadcastWorkerCount)
}

func (w *BroadcastWorker) run(ctx context.Context, id int) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("broadcast worker panic", "worker_id", id, "panic", r, "stack", string(debug.Stack()))
			time.Sleep(5 * time.Second)
			go w.run(ctx, id)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		job, err := w.queries.ClaimBroadcastJob(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(broadcastWorkerIdleSleep):
			}
			continue
		}
		if err != nil {
			slog.Error("broadcast worker claim failed", "worker_id", id, "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(broadcastWorkerIdleSleep):
			}
			continue
		}

		w.processJob(ctx, job)
	}
}

func (w *BroadcastWorker) processJob(ctx context.Context, job db.BroadcastJob) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("broadcast job panic", "job_id", job.ID, "panic", r, "stack", string(debug.Stack()))
			w.scheduleRetry(ctx, job, "panic during send")
		}
	}()

	broadcast, err := w.queries.GetBroadcast(ctx, db.GetBroadcastParams{ID: job.BroadcastID, ProjectID: job.ProjectID})
	if err != nil {
		slog.Error("broadcast job parent lookup failed", "job_id", job.ID, "error", err)
		_ = w.queries.MarkBroadcastJobFailed(ctx, db.MarkBroadcastJobFailedParams{
			ID:        job.ID,
			LastError: "parent broadcast not found",
		})
		_ = w.queries.IncrementBroadcastFailed(ctx, job.BroadcastID)
		w.checkBroadcastCompletion(ctx, job.BroadcastID)
		return
	}

	outcome, sendErr := w.emailService.SendBroadcastJob(ctx, job, broadcast)

	switch outcome {
	case BroadcastJobOutcomeSent:
		_ = w.queries.MarkBroadcastJobSent(ctx, job.ID)
		_ = w.queries.IncrementBroadcastSent(ctx, job.BroadcastID)

	case BroadcastJobOutcomeSuppressed:
		_ = w.queries.MarkBroadcastJobSuppressed(ctx, job.ID)
		_ = w.queries.IncrementBroadcastSuppressed(ctx, job.BroadcastID)

	case BroadcastJobOutcomeBounced:
		errMsg := errString(sendErr)
		_ = w.queries.MarkBroadcastJobBounced(ctx, db.MarkBroadcastJobBouncedParams{
			ID:        job.ID,
			LastError: errMsg,
		})
		_ = w.queries.IncrementBroadcastFailed(ctx, job.BroadcastID)

	case BroadcastJobOutcomeTransientError:
		errMsg := errString(sendErr)
		if int(job.Attempts) >= broadcastJobMaxAttempts {
			_ = w.queries.MarkBroadcastJobFailed(ctx, db.MarkBroadcastJobFailedParams{
				ID:        job.ID,
				LastError: errMsg,
			})
			_ = w.queries.IncrementBroadcastFailed(ctx, job.BroadcastID)
			slog.Error("broadcast job failed permanently", "job_id", job.ID, "attempts", job.Attempts, "error", errMsg)
		} else {
			w.scheduleRetry(ctx, job, errMsg)
			return
		}
	}

	w.checkBroadcastCompletion(ctx, job.BroadcastID)
}

func (w *BroadcastWorker) checkBroadcastCompletion(ctx context.Context, broadcastID uuid.UUID) {
	remaining, err := w.queries.CountBroadcastJobsRemaining(ctx, broadcastID)
	if err != nil {
		slog.Error("broadcast completion check failed", "broadcast_id", broadcastID, "error", err)
		return
	}
	if remaining > 0 {
		return
	}
	if err := w.queries.MarkBroadcastCompleted(ctx, broadcastID); err != nil {
		slog.Error("broadcast mark completed failed", "broadcast_id", broadcastID, "error", err)
		return
	}
	if err := w.queries.MarkCampaignDoneFromBroadcast(ctx, broadcastID); err != nil {
		slog.Error("broadcast linked-campaign update failed", "broadcast_id", broadcastID, "error", err)
	}
	slog.Info("broadcast completed", "broadcast_id", broadcastID)
}

func (w *BroadcastWorker) scheduleRetry(ctx context.Context, job db.BroadcastJob, errMsg string) {
	backoff := backoffDuration(int(job.Attempts))
	nextAt := time.Now().Add(backoff)
	if err := w.queries.ScheduleBroadcastJobRetry(ctx, db.ScheduleBroadcastJobRetryParams{
		ID:          job.ID,
		ScheduledAt: nextAt,
		LastError:   errMsg,
	}); err != nil {
		slog.Error("broadcast job retry schedule failed", "job_id", job.ID, "error", err)
		return
	}
	slog.Warn("broadcast job retry scheduled", "job_id", job.ID, "backoff", backoff.String(), "attempt", job.Attempts, "error", errMsg)
}

func backoffDuration(attempts int) time.Duration {
	switch attempts {
	case 1:
		return 30 * time.Second
	case 2:
		return 2 * time.Minute
	case 3:
		return 8 * time.Minute
	case 4:
		return 30 * time.Minute
	default:
		return 1 * time.Hour
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
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
		log.Printf("broadcast worker: reset stuck jobs failed: %v", err)
	} else if rows > 0 {
		log.Printf("broadcast worker: reset %d stuck 'sending' jobs to retry", rows)
	}

	for i := 0; i < broadcastWorkerCount; i++ {
		go w.run(ctx, i)
	}
	log.Printf("Broadcast worker started (%d goroutines)", broadcastWorkerCount)
}

func (w *BroadcastWorker) run(ctx context.Context, id int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("broadcast worker %d: panic: %v\n%s", id, r, debug.Stack())
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
			log.Printf("broadcast worker %d: claim failed: %v", id, err)
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
			log.Printf("broadcast job %s: panic: %v\n%s", job.ID, r, debug.Stack())
			w.scheduleRetry(ctx, job, "panic during send")
		}
	}()

	broadcast, err := w.queries.GetBroadcast(ctx, db.GetBroadcastParams{ID: job.BroadcastID, ProjectID: job.ProjectID})
	if err != nil {
		log.Printf("broadcast job %s: parent broadcast lookup failed: %v", job.ID, err)
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
			log.Printf("broadcast job %s: failed after %d attempts: %s", job.ID, job.Attempts, errMsg)
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
		log.Printf("broadcast %s: completion check failed: %v", broadcastID, err)
		return
	}
	if remaining > 0 {
		return
	}
	if err := w.queries.MarkBroadcastCompleted(ctx, broadcastID); err != nil {
		log.Printf("broadcast %s: mark completed failed: %v", broadcastID, err)
		return
	}
	log.Printf("Broadcast %s completed", broadcastID)
}

func (w *BroadcastWorker) scheduleRetry(ctx context.Context, job db.BroadcastJob, errMsg string) {
	backoff := backoffDuration(int(job.Attempts))
	nextAt := time.Now().Add(backoff)
	if err := w.queries.ScheduleBroadcastJobRetry(ctx, db.ScheduleBroadcastJobRetryParams{
		ID:          job.ID,
		ScheduledAt: nextAt,
		LastError:   errMsg,
	}); err != nil {
		log.Printf("broadcast job %s: retry schedule failed: %v", job.ID, err)
		return
	}
	log.Printf("broadcast job %s: retry scheduled in %s (attempt %d, error: %s)", job.ID, backoff, job.Attempts, errMsg)
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

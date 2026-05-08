-- name: BulkInsertBroadcastJobs :exec
INSERT INTO broadcast_jobs (broadcast_id, project_id, subscriber_id, recipient_email)
SELECT
    @broadcast_id::uuid,
    @project_id::uuid,
    (j->>'subscriber_id')::uuid,
    j->>'recipient_email'
FROM jsonb_array_elements(@jobs::jsonb) AS j;

-- name: ClaimBroadcastJob :one
UPDATE broadcast_jobs SET
    status = 'sending',
    attempts = attempts + 1
WHERE id = (
    SELECT id FROM broadcast_jobs
    WHERE status IN ('pending', 'retry')
      AND scheduled_at <= NOW()
    ORDER BY scheduled_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: MarkBroadcastJobSent :exec
UPDATE broadcast_jobs SET
    status = 'sent',
    completed_at = NOW(),
    last_error = NULL
WHERE id = $1;

-- name: MarkBroadcastJobFailed :exec
UPDATE broadcast_jobs SET
    status = 'failed',
    completed_at = NOW(),
    last_error = @last_error::text
WHERE id = @id;

-- name: MarkBroadcastJobSuppressed :exec
UPDATE broadcast_jobs SET
    status = 'suppressed',
    completed_at = NOW()
WHERE id = $1;

-- name: MarkBroadcastJobBounced :exec
UPDATE broadcast_jobs SET
    status = 'bounced',
    completed_at = NOW(),
    last_error = @last_error::text
WHERE id = @id;

-- name: ScheduleBroadcastJobRetry :exec
UPDATE broadcast_jobs SET
    status = 'retry',
    scheduled_at = @scheduled_at,
    last_error = @last_error::text
WHERE id = @id;

-- name: CountBroadcastJobsRemaining :one
SELECT COUNT(*) FROM broadcast_jobs
WHERE broadcast_id = $1
  AND status IN ('pending', 'retry', 'sending');

-- name: ResetStuckSendingJobs :execrows
UPDATE broadcast_jobs SET
    status = 'retry',
    scheduled_at = NOW()
WHERE status = 'sending';

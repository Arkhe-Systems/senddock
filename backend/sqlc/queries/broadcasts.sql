-- name: CreateBroadcast :one
INSERT INTO broadcasts (project_id, template_id, subject, variables, html_fields, total_recipients)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: IncrementBroadcastSent :exec
UPDATE broadcasts SET sent_count = sent_count + 1 WHERE id = $1;

-- name: IncrementBroadcastFailed :exec
UPDATE broadcasts SET failed_count = failed_count + 1 WHERE id = $1;

-- name: IncrementBroadcastSuppressed :exec
UPDATE broadcasts SET suppressed_count = suppressed_count + 1 WHERE id = $1;

-- name: MarkBroadcastCompleted :exec
UPDATE broadcasts SET status = 'completed', finished_at = NOW() WHERE id = $1;

-- name: ReconcileStuckBroadcasts :exec
-- Marks 'completed' any broadcast (status 'sending' or legacy 'interrupted')
-- whose per-recipient jobs have all settled. In-flight broadcasts whose jobs
-- the worker will pick up via ResetStuckSendingJobs stay 'sending'.
UPDATE broadcasts b
SET status = 'completed',
    finished_at = COALESCE(b.finished_at, NOW())
WHERE b.status IN ('sending', 'interrupted')
  AND NOT EXISTS (
    SELECT 1 FROM broadcast_jobs j
    WHERE j.broadcast_id = b.id
      AND j.status IN ('pending', 'retry', 'sending')
  );

-- name: ListBroadcastsByProject :many
SELECT * FROM broadcasts
WHERE project_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: CountBroadcastsByProject :one
SELECT COUNT(*) FROM broadcasts WHERE project_id = $1;

-- name: GetBroadcast :one
SELECT * FROM broadcasts WHERE id = $1 AND project_id = $2;

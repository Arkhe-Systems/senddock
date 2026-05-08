-- name: CreateBroadcast :one
INSERT INTO broadcasts (project_id, template_id, subject, variables, total_recipients)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: IncrementBroadcastSent :exec
UPDATE broadcasts SET sent_count = sent_count + 1 WHERE id = $1;

-- name: IncrementBroadcastFailed :exec
UPDATE broadcasts SET failed_count = failed_count + 1 WHERE id = $1;

-- name: IncrementBroadcastSuppressed :exec
UPDATE broadcasts SET suppressed_count = suppressed_count + 1 WHERE id = $1;

-- name: MarkBroadcastCompleted :exec
UPDATE broadcasts SET status = 'completed', finished_at = NOW() WHERE id = $1;

-- name: MarkInProgressBroadcastsInterrupted :exec
UPDATE broadcasts SET status = 'interrupted', finished_at = NOW() WHERE status = 'sending';

-- name: ListBroadcastsByProject :many
SELECT * FROM broadcasts
WHERE project_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: CountBroadcastsByProject :one
SELECT COUNT(*) FROM broadcasts WHERE project_id = $1;

-- name: GetBroadcast :one
SELECT * FROM broadcasts WHERE id = $1 AND project_id = $2;

-- name: CreateEmailLog :one
INSERT INTO email_logs (project_id, subscriber_id, template_id, to_email, subject, status, error)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListEmailLogsByProject :many
SELECT * FROM email_logs
WHERE project_id = $1
ORDER BY sent_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEmailLogsByProject :one
SELECT COUNT(*) FROM email_logs WHERE project_id = $1;

-- name: CountEmailLogsByStatus :one
SELECT COUNT(*) FROM email_logs WHERE project_id = $1 AND status = $2;

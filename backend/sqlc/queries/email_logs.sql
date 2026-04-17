-- name: CreateEmailLog :one
INSERT INTO email_logs (project_id, subscriber_id, template_id, to_email, subject, status, error)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListEmailLogsByProject :many
SELECT * FROM email_logs
WHERE project_id = $1
ORDER BY sent_at DESC
LIMIT $2 OFFSET $3;

-- name: ListEmailLogsByProjectFiltered :many
SELECT * FROM email_logs
WHERE project_id = $1
AND ($4::text = '' OR status = $4::text)
AND ($5::timestamptz = '0001-01-01'::timestamptz OR sent_at >= $5)
AND ($6::timestamptz = '0001-01-01'::timestamptz OR sent_at <= $6)
ORDER BY sent_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEmailLogsByProjectFiltered :one
SELECT COUNT(*) FROM email_logs
WHERE project_id = $1
AND ($2::text = '' OR status = $2::text)
AND ($3::timestamptz = '0001-01-01'::timestamptz OR sent_at >= $3)
AND ($4::timestamptz = '0001-01-01'::timestamptz OR sent_at <= $4);

-- name: CountEmailLogsByProject :one
SELECT COUNT(*) FROM email_logs WHERE project_id = $1;

-- name: CountEmailLogsByStatus :one
SELECT COUNT(*) FROM email_logs WHERE project_id = $1 AND status = $2;

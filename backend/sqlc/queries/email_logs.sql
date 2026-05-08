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
AND ($7::text = '' OR to_email ILIKE '%' || $7::text || '%' OR subject ILIKE '%' || $7::text || '%')
AND ($8::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR template_id = $8::uuid)
ORDER BY sent_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEmailLogsByProjectFiltered :one
SELECT COUNT(*) FROM email_logs
WHERE project_id = $1
AND ($2::text = '' OR status = $2::text)
AND ($3::timestamptz = '0001-01-01'::timestamptz OR sent_at >= $3)
AND ($4::timestamptz = '0001-01-01'::timestamptz OR sent_at <= $4)
AND ($5::text = '' OR to_email ILIKE '%' || $5::text || '%' OR subject ILIKE '%' || $5::text || '%')
AND ($6::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR template_id = $6::uuid);

-- name: ListEmailLogsByProjectExport :many
SELECT * FROM email_logs
WHERE project_id = $1
AND ($2::text = '' OR status = $2::text)
AND ($3::timestamptz = '0001-01-01'::timestamptz OR sent_at >= $3)
AND ($4::timestamptz = '0001-01-01'::timestamptz OR sent_at <= $4)
AND ($5::text = '' OR to_email ILIKE '%' || $5::text || '%' OR subject ILIKE '%' || $5::text || '%')
AND ($6::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR template_id = $6::uuid)
ORDER BY sent_at DESC;

-- name: CountEmailLogsByProject :one
SELECT COUNT(*) FROM email_logs WHERE project_id = $1;

-- name: CountEmailLogsByStatus :one
SELECT COUNT(*) FROM email_logs WHERE project_id = $1 AND status = $2;

-- name: MarkEmailOpened :execrows
UPDATE email_logs SET opened_at = NOW() WHERE id = $1 AND opened_at IS NULL;

-- name: CountEmailLogsOpened :one
SELECT COUNT(*) FROM email_logs WHERE project_id = $1 AND opened_at IS NOT NULL;

-- name: MarkEmailClicked :execrows
UPDATE email_logs SET clicked_at = NOW() WHERE id = $1 AND clicked_at IS NULL;

-- name: CountEmailLogsClicked :one
SELECT COUNT(*) FROM email_logs WHERE project_id = $1 AND clicked_at IS NOT NULL;

-- name: CreateEmailClick :exec
INSERT INTO email_clicks (log_id, url, url_hash, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5);

-- name: GetEmailLogProjectID :one
SELECT project_id FROM email_logs WHERE id = $1;

-- name: UpdateEmailLogStatus :exec
UPDATE email_logs SET status = $2, error = $3 WHERE id = $1;

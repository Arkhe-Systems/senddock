-- name: CreateTemplate :one
INSERT INTO templates (project_id, name, subject, html_body, text_body)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTemplateByID :one
SELECT * FROM templates WHERE id = $1 AND project_id = $2;

-- name: ListTemplatesByProject :many
SELECT * FROM templates
WHERE project_id = $1
ORDER BY updated_at DESC;

-- name: UpdateTemplate :one
UPDATE templates SET
    name = $3,
    subject = $4,
    html_body = $5,
    text_body = $6,
    updated_at = NOW()
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: DeleteTemplate :exec
DELETE FROM templates WHERE id = $1 AND project_id = $2;

-- name: CountTemplatesByProject :one
SELECT COUNT(*) FROM templates WHERE project_id = $1;

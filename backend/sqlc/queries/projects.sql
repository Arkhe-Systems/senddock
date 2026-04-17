-- name: CreateProject :one
INSERT INTO projects (user_id, name, description, from_name, from_email)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProjectsByUserID :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1 AND user_id = $2;

-- name: GetProjectByIDOnly :one
SELECT * FROM projects WHERE id = $1;

-- name: UpdateProject :one
UPDATE projects SET
    name = $3,
    description = $4,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateProjectSMTP :one
UPDATE projects SET
    smtp_host = $3,
    smtp_port = $4,
    smtp_user = $5,
    smtp_password_encrypted = $6,
    from_name = $7,
    from_email = $8,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1 AND user_id = $2;

-- name: CountProjectsByUserID :one
SELECT COUNT(*) FROM projects WHERE user_id = $1;
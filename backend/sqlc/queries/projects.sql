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

-- name: RotateBounceToken :one
UPDATE projects SET bounce_token = gen_random_uuid(), updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetProjectByBounceToken :one
SELECT * FROM projects WHERE id = $1 AND bounce_token = $2;

-- name: UpdateBounceIMAP :one
UPDATE projects SET
    bounce_imap_host = $3,
    bounce_imap_port = $4,
    bounce_imap_user = $5,
    bounce_imap_password_encrypted = $6,
    bounce_imap_folder = $7,
    bounce_imap_enabled = $8,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: ListProjectsWithBounceIMAP :many
SELECT * FROM projects
WHERE bounce_imap_enabled = TRUE
  AND bounce_imap_host IS NOT NULL
  AND bounce_imap_user IS NOT NULL
  AND bounce_imap_password_encrypted IS NOT NULL;
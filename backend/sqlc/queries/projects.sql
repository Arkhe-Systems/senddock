-- name: CreateProject :one
INSERT INTO projects (user_id, name, from_name, from_email)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProjectsByUserID :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1 AND user_id = $2;

-- name: UpdateProject :one
UPDATE projects SET
    name = $3,
    from_name = $4,
    from_email = $5,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1 AND user_id = $2;

-- name: CountProjectsByUserID :one
SELECT COUNT(*) FROM projects WHERE user_id = $1;
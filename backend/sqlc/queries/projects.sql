-- name: CreateProject :one
INSERT INTO projects (workspace_id, user_id, name, description, from_name, from_email)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProjectsByUserID :many
SELECT p.* FROM projects p
JOIN workspace_members m ON m.workspace_id = p.workspace_id
WHERE m.user_id = $1
ORDER BY p.created_at DESC;

-- name: GetProjectsByWorkspaceForUser :many
SELECT p.* FROM projects p
JOIN workspace_members m ON m.workspace_id = p.workspace_id
WHERE p.workspace_id = $1 AND m.user_id = $2
ORDER BY p.created_at DESC;

-- name: GetProjectByID :one
SELECT p.* FROM projects p
JOIN workspace_members m ON m.workspace_id = p.workspace_id
WHERE p.id = $1 AND m.user_id = $2;

-- name: GetProjectByIDOnly :one
SELECT * FROM projects WHERE id = $1;

-- name: UpdateProject :one
UPDATE projects SET
    name = $3,
    description = $4,
    updated_at = NOW()
WHERE id = $1
  AND workspace_id IN (SELECT wm.workspace_id FROM workspace_members wm WHERE wm.user_id = $2)
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
WHERE id = $1
  AND workspace_id IN (SELECT wm.workspace_id FROM workspace_members wm WHERE wm.user_id = $2)
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1
  AND workspace_id IN (SELECT wm.workspace_id FROM workspace_members wm WHERE wm.user_id = $2);

-- name: CountProjectsByUserID :one
SELECT COUNT(*) FROM projects p
JOIN workspace_members m ON m.workspace_id = p.workspace_id
WHERE m.user_id = $1;

-- name: RotateBounceToken :one
UPDATE projects SET bounce_token = gen_random_uuid(), updated_at = NOW()
WHERE id = $1
  AND workspace_id IN (SELECT wm.workspace_id FROM workspace_members wm WHERE wm.user_id = $2)
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
WHERE id = $1
  AND workspace_id IN (SELECT wm.workspace_id FROM workspace_members wm WHERE wm.user_id = $2)
RETURNING *;

-- name: ListProjectsWithBounceIMAP :many
SELECT * FROM projects
WHERE bounce_imap_enabled = TRUE
  AND bounce_imap_host IS NOT NULL
  AND bounce_imap_user IS NOT NULL
  AND bounce_imap_password_encrypted IS NOT NULL;

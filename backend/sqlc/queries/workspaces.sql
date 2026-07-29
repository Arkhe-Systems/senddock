-- name: CreateWorkspace :one
INSERT INTO workspaces (name, created_by)
VALUES ($1, $2)
RETURNING *;

-- name: GetWorkspaceByID :one
SELECT * FROM workspaces WHERE id = $1;

-- name: ListWorkspacesByUser :many
SELECT w.*
FROM workspaces w
JOIN workspace_members m ON m.workspace_id = w.id
WHERE m.user_id = $1
ORDER BY w.created_at ASC;

-- name: RenameWorkspace :one
UPDATE workspaces SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces WHERE id = $1;

-- name: CountProjectsInWorkspace :one
SELECT COUNT(*) FROM projects WHERE workspace_id = $1;

-- name: AddWorkspaceMember :one
INSERT INTO workspace_members (workspace_id, user_id, role, invited_by)
VALUES ($1, $2, $3, $4)
ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = EXCLUDED.role
RETURNING *;

-- name: ListWorkspaceMembers :many
SELECT m.workspace_id, m.user_id, m.role, m.invited_by, m.created_at,
       u.email AS user_email, u.name AS user_name
FROM workspace_members m
JOIN users u ON u.id = m.user_id
WHERE m.workspace_id = $1
ORDER BY m.created_at ASC;

-- name: GetWorkspaceMember :one
SELECT * FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: RemoveWorkspaceMember :exec
DELETE FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: UpdateWorkspaceMemberRole :one
UPDATE workspace_members SET role = $3
WHERE workspace_id = $1 AND user_id = $2
RETURNING *;

-- name: CountWorkspaceOwners :one
SELECT COUNT(*) FROM workspace_members
WHERE workspace_id = $1 AND role = 'owner';

-- name: IsWorkspaceMember :one
SELECT EXISTS (
    SELECT 1 FROM workspace_members
    WHERE workspace_id = $1 AND user_id = $2
);

-- name: CountOwnedWorkspacesByUser :one
SELECT COUNT(*) FROM workspace_members
WHERE user_id = $1 AND role = 'owner';

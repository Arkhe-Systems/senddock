-- name: CreateAPIKey :one
INSERT INTO api_keys (project_id, name, key_hash, key_prefix)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListAPIKeysByProject :many
SELECT * FROM api_keys
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = NOW() WHERE id = $1;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys WHERE id = $1 AND project_id = $2;

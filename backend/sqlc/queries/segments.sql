-- name: CreateSegment :one
INSERT INTO segments (project_id, name, predicate)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListSegmentsByProject :many
SELECT * FROM segments
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: GetSegment :one
SELECT * FROM segments
WHERE id = $1 AND project_id = $2;

-- name: UpdateSegment :one
UPDATE segments SET
    name = $3,
    predicate = $4,
    updated_at = NOW()
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: DeleteSegment :exec
DELETE FROM segments
WHERE id = $1 AND project_id = $2;

-- name: UpsertSuppression :one
INSERT INTO suppressions (project_id, email_normalized, reason, source)
VALUES ($1, $2, $3, $4)
ON CONFLICT (project_id, email_normalized) DO UPDATE
    SET last_seen_at = NOW(),
        reason = EXCLUDED.reason,
        source = COALESCE(EXCLUDED.source, suppressions.source)
RETURNING *;

-- name: IsSuppressed :one
SELECT EXISTS (
    SELECT 1 FROM suppressions
    WHERE project_id = $1 AND email_normalized = LOWER($2)
);

-- name: ListSuppressionsByProject :many
SELECT * FROM suppressions
WHERE project_id = $1
  AND ($2::text = '' OR reason = $2::text)
ORDER BY last_seen_at DESC
LIMIT $3 OFFSET $4;

-- name: CountSuppressionsByProject :one
SELECT COUNT(*) FROM suppressions
WHERE project_id = $1
  AND ($2::text = '' OR reason = $2::text);

-- name: DeleteSuppression :exec
DELETE FROM suppressions
WHERE id = $1 AND project_id = $2;

-- name: DeleteSuppressionByEmail :exec
DELETE FROM suppressions
WHERE project_id = $1 AND email_normalized = LOWER($2);

-- name: GetInstanceSettings :one
SELECT * FROM instance_settings WHERE id = true;

-- name: UpdateInstanceSettings :one
UPDATE instance_settings SET
    public_url = $1,
    session_idle_timeout_minutes = $2,
    updated_at = NOW()
WHERE id = true
RETURNING *;

-- name: SetInstancePublicURL :one
UPDATE instance_settings SET
    public_url = $1,
    updated_at = NOW()
WHERE id = true
RETURNING *;

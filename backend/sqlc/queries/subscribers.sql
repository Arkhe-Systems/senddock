-- name: CreateSubscriber :one
INSERT INTO subscribers (project_id, email, name, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSubscriberByID :one
SELECT * FROM subscribers WHERE id = $1 AND project_id = $2;

-- name: GetSubscriberByEmail :one
SELECT * FROM subscribers WHERE email = $1 AND project_id = $2;

-- name: ListSubscribersByProject :many
SELECT * FROM subscribers
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveSubscribersByProject :many
SELECT * FROM subscribers
WHERE project_id = $1 AND status = 'active'
ORDER BY created_at DESC;

-- name: CountSubscribersByProject :one
SELECT COUNT(*) FROM subscribers WHERE project_id = $1;

-- name: CountActiveSubscribersByProject :one
SELECT COUNT(*) FROM subscribers WHERE project_id = $1 AND status = 'active';

-- name: UpdateSubscriberStatus :one
UPDATE subscribers SET
    status = $3,
    unsubscribed_at = CASE WHEN $3::text = 'unsubscribed' THEN NOW() ELSE unsubscribed_at END,
    updated_at = NOW()
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: UpdateSubscriber :one
UPDATE subscribers SET
    name = $3,
    email = $4,
    updated_at = NOW()
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: DeleteSubscriber :exec
DELETE FROM subscribers WHERE id = $1 AND project_id = $2;

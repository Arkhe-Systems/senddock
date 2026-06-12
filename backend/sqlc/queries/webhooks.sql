-- name: CreateWebhook :one
INSERT INTO webhooks (project_id, url, secret, events)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListWebhooksByProject :many
SELECT * FROM webhooks
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: GetWebhook :one
SELECT * FROM webhooks
WHERE id = $1 AND project_id = $2;

-- name: GetWebhookForDispatchInternal :one
-- INTERNAL USE ONLY. Lookup by webhook id without tenant scope.
-- Only safe to call from the webhook dispatcher worker where the webhook id
-- already comes from a verified webhook_deliveries row in the same project.
-- NEVER call from a user-facing handler.
SELECT * FROM webhooks
WHERE id = $1;

-- name: DeleteWebhook :exec
DELETE FROM webhooks
WHERE id = $1 AND project_id = $2;

-- name: ListActiveWebhooksForEvent :many
SELECT * FROM webhooks
WHERE project_id = sqlc.arg(project_id) AND active AND sqlc.arg(event_type)::text = ANY(events);

-- name: CreateWebhookDelivery :one
INSERT INTO webhook_deliveries (webhook_id, event_type, payload)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ClaimPendingDeliveries :many
UPDATE webhook_deliveries
SET status = 'inflight'
WHERE id IN (
    SELECT id FROM webhook_deliveries
    WHERE status = 'pending' AND next_attempt_at <= NOW()
    ORDER BY next_attempt_at
    FOR UPDATE SKIP LOCKED
    LIMIT $1
)
RETURNING *;

-- name: MarkDeliverySuccess :exec
UPDATE webhook_deliveries
SET status = 'delivered',
    attempts = attempts + 1,
    last_status_code = $2,
    delivered_at = NOW()
WHERE id = $1;

-- name: MarkDeliveryRetry :exec
UPDATE webhook_deliveries
SET status = 'pending',
    attempts = attempts + 1,
    last_status_code = $2,
    last_error = $3,
    next_attempt_at = $4
WHERE id = $1;

-- name: MarkDeliveryFailed :exec
UPDATE webhook_deliveries
SET status = 'failed',
    attempts = attempts + 1,
    last_status_code = $2,
    last_error = $3
WHERE id = $1;

-- name: ListWebhookDeliveries :many
SELECT * FROM webhook_deliveries
WHERE webhook_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

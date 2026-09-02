-- name: CreateSubscriber :one
INSERT INTO subscribers (project_id, email, name, status, metadata, tags)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSubscriberByID :one
SELECT * FROM subscribers WHERE id = $1 AND project_id = $2;

-- name: GetSubscriberByEmail :one
SELECT * FROM subscribers WHERE email = $1 AND project_id = $2;

-- name: ListSubscribersByProject :many
SELECT s.* FROM subscribers s
WHERE s.project_id = $1
AND ($4::text = '' OR s.status = $4::text)
AND ($5::text = '' OR $5::text = ANY(s.tags))
AND ($6::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR s.id IN (
    SELECT ns.subscriber_id FROM newsletter_subscriptions ns
    WHERE ns.newsletter_id = $6::uuid AND ns.unsubscribed_at IS NULL))
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveSubscribersByProject :many
SELECT * FROM subscribers
WHERE project_id = $1 AND status = 'active'
ORDER BY created_at DESC;

-- name: CountSubscribersByProject :one
SELECT COUNT(*) FROM subscribers s
WHERE s.project_id = $1
AND ($2::text = '' OR s.status = $2::text)
AND ($3::text = '' OR $3::text = ANY(s.tags))
AND ($4::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR s.id IN (
    SELECT ns.subscriber_id FROM newsletter_subscriptions ns
    WHERE ns.newsletter_id = $4::uuid AND ns.unsubscribed_at IS NULL));

-- name: CountActiveSubscribersByProject :one
SELECT COUNT(*) FROM subscribers WHERE project_id = $1 AND status = 'active';

-- name: UpdateSubscriberStatus :one
UPDATE subscribers SET
    status = $3,
    unsubscribed_at = CASE WHEN $4 = 'unsubscribed' THEN NOW() ELSE unsubscribed_at END,
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

-- name: UpdateSubscriberMetadata :one
UPDATE subscribers SET
    metadata = $3,
    updated_at = NOW()
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: SetSubscriberTags :one
UPDATE subscribers SET
    tags = $3,
    updated_at = NOW()
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: ListDistinctTagsByProject :many
SELECT DISTINCT unnest(tags)::text AS tag
FROM subscribers
WHERE project_id = $1
ORDER BY tag;

-- name: BulkAddSubscriberTags :exec
UPDATE subscribers SET
    tags = ARRAY(SELECT DISTINCT unnest(tags || $3::text[])),
    updated_at = NOW()
WHERE project_id = $1 AND id = ANY($2::uuid[]);

-- name: BulkRemoveSubscriberTags :exec
UPDATE subscribers SET
    tags = ARRAY(SELECT unnest(tags) EXCEPT SELECT unnest($3::text[])),
    updated_at = NOW()
WHERE project_id = $1 AND id = ANY($2::uuid[]);

-- name: DeleteSubscriber :exec
DELETE FROM subscribers WHERE id = $1 AND project_id = $2;

-- name: BulkDeleteSubscribers :exec
DELETE FROM subscribers 
WHERE project_id = $1 AND id = ANY($2::uuid[]);

-- name: BulkUpdateSubscriberStatus :exec
UPDATE subscribers 
SET 
    status = $3,
    unsubscribed_at = CASE WHEN $3 = 'unsubscribed' THEN NOW() ELSE unsubscribed_at END,
    updated_at = NOW()
WHERE project_id = $1 AND id = ANY($2::uuid[]);

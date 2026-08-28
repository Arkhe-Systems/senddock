-- name: CreateNewsletter :one
INSERT INTO newsletters (project_id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetNewsletterByID :one
SELECT * FROM newsletters
WHERE id = $1 AND project_id = $2;

-- name: ListNewslettersByProject :many
SELECT n.*,
    (SELECT COUNT(*) FROM newsletter_subscriptions ns
       JOIN subscribers s ON s.id = ns.subscriber_id
      WHERE ns.newsletter_id = n.id
        AND ns.unsubscribed_at IS NULL
        AND s.status = 'active') AS active_count
FROM newsletters n
WHERE n.project_id = $1
ORDER BY n.created_at DESC;

-- name: UpdateNewsletter :one
UPDATE newsletters SET
    name = $3,
    description = $4,
    updated_at = NOW()
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: DeleteNewsletter :exec
DELETE FROM newsletters
WHERE id = $1 AND project_id = $2;

-- name: ListActiveNewsletterSubscribers :many
SELECT s.id, s.email FROM subscribers s
JOIN newsletter_subscriptions ns ON ns.subscriber_id = s.id
WHERE ns.newsletter_id = $1
  AND s.project_id = $2
  AND s.status = 'active'
  AND ns.unsubscribed_at IS NULL
ORDER BY s.created_at DESC;

-- name: ListSubscriberNewsletters :many
SELECT n.id, n.name, ns.unsubscribed_at FROM newsletters n
JOIN newsletter_subscriptions ns ON ns.newsletter_id = n.id
WHERE ns.subscriber_id = $1 AND n.project_id = $2
ORDER BY n.created_at DESC;

-- name: AddNewsletterSubscription :exec
INSERT INTO newsletter_subscriptions (newsletter_id, subscriber_id, project_id)
SELECT n.id, sqlc.arg(subscriber_id)::uuid, n.project_id
FROM newsletters n
WHERE n.id = sqlc.arg(newsletter_id)::uuid AND n.project_id = sqlc.arg(project_id)::uuid
ON CONFLICT (newsletter_id, subscriber_id) DO UPDATE SET unsubscribed_at = NULL;

-- name: DeleteSubscriberNewsletterSubscriptions :exec
DELETE FROM newsletter_subscriptions
WHERE subscriber_id = $1 AND project_id = $2;

-- name: MarkNewsletterUnsubscribed :exec
INSERT INTO newsletter_subscriptions (newsletter_id, subscriber_id, project_id, unsubscribed_at)
SELECT n.id, sqlc.arg(subscriber_id)::uuid, n.project_id, NOW()
FROM newsletters n
WHERE n.id = sqlc.arg(newsletter_id)::uuid AND n.project_id = sqlc.arg(project_id)::uuid
ON CONFLICT (newsletter_id, subscriber_id) DO UPDATE SET unsubscribed_at = NOW();

-- name: BulkAddNewsletterSubscriptions :exec
INSERT INTO newsletter_subscriptions (newsletter_id, subscriber_id, project_id)
SELECT n.id, s.id, n.project_id
FROM newsletters n
CROSS JOIN subscribers s
WHERE n.id = sqlc.arg(newsletter_id)::uuid
  AND n.project_id = sqlc.arg(project_id)::uuid
  AND s.project_id = sqlc.arg(project_id)::uuid
  AND s.id = ANY(sqlc.arg(subscriber_ids)::uuid[])
ON CONFLICT (newsletter_id, subscriber_id) DO UPDATE SET unsubscribed_at = NULL;

-- name: BulkRemoveNewsletterSubscriptions :exec
DELETE FROM newsletter_subscriptions
WHERE newsletter_id = sqlc.arg(newsletter_id)::uuid
  AND project_id = sqlc.arg(project_id)::uuid
  AND subscriber_id = ANY(sqlc.arg(subscriber_ids)::uuid[]);

-- name: CreateCampaign :one
INSERT INTO campaigns (project_id, template_id, name, subject, scheduled_at, variables, newsletter_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListCampaignsByProject :many
SELECT
    c.id, c.project_id, c.template_id, c.name, c.subject,
    c.scheduled_at, c.sent_at, c.created_at, c.status,
    COALESCE(b.sent_count, c.sent_count)::int AS sent_count,
    COALESCE(b.failed_count, c.failed_count)::int AS failed_count,
    c.variables,
    c.broadcast_id
FROM campaigns c
LEFT JOIN broadcasts b ON b.id = c.broadcast_id
WHERE c.project_id = $1
ORDER BY c.created_at DESC;

-- name: GetCampaignByID :one
SELECT
    c.id, c.project_id, c.template_id, c.name, c.subject,
    c.scheduled_at, c.sent_at, c.created_at, c.status,
    COALESCE(b.sent_count, c.sent_count)::int AS sent_count,
    COALESCE(b.failed_count, c.failed_count)::int AS failed_count,
    c.variables,
    c.broadcast_id
FROM campaigns c
LEFT JOIN broadcasts b ON b.id = c.broadcast_id
WHERE c.id = $1 AND c.project_id = $2;

-- name: GetPendingCampaigns :many
SELECT * FROM campaigns
WHERE status = 'scheduled' AND scheduled_at <= NOW()
ORDER BY scheduled_at ASC;

-- name: UpdateCampaignStatus :exec
UPDATE campaigns SET
    status = @status::text,
    sent_at = CASE WHEN @status::text = 'sent' THEN NOW() ELSE sent_at END,
    sent_count = @sent_count,
    failed_count = @failed_count
WHERE id = @id;

-- name: ClaimCampaignForExecution :execrows
UPDATE campaigns SET status = 'sending'
WHERE id = $1 AND status = 'scheduled';

-- name: SetCampaignBroadcast :exec
UPDATE campaigns SET broadcast_id = @broadcast_id WHERE id = @id;

-- name: MarkCampaignDoneFromBroadcast :exec
UPDATE campaigns
SET status = 'sent',
    sent_count = b.sent_count,
    failed_count = b.failed_count,
    sent_at = NOW()
FROM broadcasts b
WHERE campaigns.broadcast_id = b.id
  AND b.id = @broadcast_id
  AND campaigns.status = 'sending';

-- name: DeleteCampaign :execrows
DELETE FROM campaigns WHERE id = $1 AND project_id = $2;

-- name: UpdateCampaign :one
UPDATE campaigns SET
    name = $3,
    template_id = $4,
    subject = $5,
    scheduled_at = $6,
    variables = $7,
    newsletter_id = $8
WHERE id = $1 AND project_id = $2 AND status = 'scheduled'
RETURNING *;

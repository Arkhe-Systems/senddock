-- name: CreateCampaign :one
INSERT INTO campaigns (project_id, template_id, name, subject, scheduled_at, variables)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListCampaignsByProject :many
SELECT * FROM campaigns
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: GetCampaignByID :one
SELECT * FROM campaigns WHERE id = $1 AND project_id = $2;

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

-- name: DeleteCampaign :exec
DELETE FROM campaigns WHERE id = $1 AND project_id = $2 AND status = 'scheduled';

-- name: UpdateCampaign :one
UPDATE campaigns SET
    name = $3,
    template_id = $4,
    subject = $5,
    scheduled_at = $6,
    variables = $7
WHERE id = $1 AND project_id = $2 AND status = 'scheduled'
RETURNING *;

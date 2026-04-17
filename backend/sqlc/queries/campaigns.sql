-- name: CreateCampaign :one
INSERT INTO campaigns (project_id, template_id, name, scheduled_at)
VALUES ($1, $2, $3, $4)
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
    status = $2,
    sent_at = CASE WHEN $2 = 'sent' THEN NOW() ELSE sent_at END,
    sent_count = $3,
    failed_count = $4
WHERE id = $1;

-- name: DeleteCampaign :exec
DELETE FROM campaigns WHERE id = $1 AND project_id = $2 AND status = 'scheduled';

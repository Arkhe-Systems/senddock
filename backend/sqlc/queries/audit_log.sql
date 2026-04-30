-- name: CreateAuditEntry :one
INSERT INTO audit_log (project_id, user_id, action, target_type, target_id, metadata, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- +goose Up
-- Ownership lease for broadcast jobs. Startup recovery resets rows left in 'sending',
-- and without an owner it also resets jobs another instance is still delivering, which
-- sends those recipients the same email twice.
ALTER TABLE broadcast_jobs ADD COLUMN worker_id UUID;
ALTER TABLE broadcast_jobs ADD COLUMN lease_expires_at TIMESTAMPTZ;

CREATE INDEX idx_broadcast_jobs_lease_expiry
    ON broadcast_jobs (lease_expires_at)
    WHERE status = 'sending';

-- +goose Down
DROP INDEX idx_broadcast_jobs_lease_expiry;
ALTER TABLE broadcast_jobs DROP COLUMN lease_expires_at;
ALTER TABLE broadcast_jobs DROP COLUMN worker_id;

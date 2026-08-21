-- +goose Up
CREATE INDEX idx_broadcast_jobs_active
    ON broadcast_jobs (status)
    WHERE status IN ('pending', 'retry', 'sending');

-- +goose Down
DROP INDEX idx_broadcast_jobs_active;

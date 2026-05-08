-- +goose Up
CREATE TABLE broadcast_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    broadcast_id UUID NOT NULL REFERENCES broadcasts(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    recipient_email TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_broadcast_jobs_pickup
    ON broadcast_jobs (scheduled_at)
    WHERE status IN ('pending', 'retry');

CREATE INDEX idx_broadcast_jobs_broadcast
    ON broadcast_jobs (broadcast_id);

-- +goose Down
DROP TABLE broadcast_jobs;

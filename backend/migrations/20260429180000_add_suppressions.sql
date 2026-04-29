-- +goose Up
CREATE TABLE suppressions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    email_normalized VARCHAR(254) NOT NULL,
    reason VARCHAR(32) NOT NULL,
    source VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX suppressions_project_email_idx ON suppressions(project_id, email_normalized);
CREATE INDEX suppressions_reason_idx ON suppressions(project_id, reason);

INSERT INTO suppressions (project_id, email_normalized, reason, source)
SELECT project_id, LOWER(email), 'unsubscribe', 'backfill: existing unsubscribed subscriber'
FROM subscribers
WHERE status = 'unsubscribed'
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE suppressions;

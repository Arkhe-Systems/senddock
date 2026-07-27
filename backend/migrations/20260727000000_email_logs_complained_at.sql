-- +goose Up
ALTER TABLE email_logs ADD COLUMN complained_at TIMESTAMPTZ;
CREATE INDEX idx_email_logs_complained_at ON email_logs(project_id, complained_at) WHERE complained_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_email_logs_complained_at;
ALTER TABLE email_logs DROP COLUMN complained_at;

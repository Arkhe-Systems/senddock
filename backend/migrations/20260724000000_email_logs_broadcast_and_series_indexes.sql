-- +goose Up
ALTER TABLE email_logs ADD COLUMN broadcast_id UUID REFERENCES broadcasts(id) ON DELETE SET NULL;

CREATE INDEX idx_email_logs_broadcast ON email_logs(broadcast_id) WHERE broadcast_id IS NOT NULL;
CREATE INDEX idx_email_logs_opened_at ON email_logs(project_id, opened_at) WHERE opened_at IS NOT NULL;
CREATE INDEX idx_email_logs_clicked_at ON email_logs(project_id, clicked_at) WHERE clicked_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_email_logs_clicked_at;
DROP INDEX IF EXISTS idx_email_logs_opened_at;
DROP INDEX IF EXISTS idx_email_logs_broadcast;
ALTER TABLE email_logs DROP COLUMN broadcast_id;

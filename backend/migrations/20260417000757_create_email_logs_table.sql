-- +goose Up
CREATE TABLE email_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    subscriber_id UUID REFERENCES subscribers(id) ON DELETE SET NULL,
    template_id UUID REFERENCES templates(id) ON DELETE SET NULL,
    to_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'sent',
    error TEXT,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_logs_project_id ON email_logs(project_id);
CREATE INDEX idx_email_logs_status ON email_logs(project_id, status);
CREATE INDEX idx_email_logs_sent_at ON email_logs(project_id, sent_at DESC);

-- +goose Down
DROP TABLE email_logs;

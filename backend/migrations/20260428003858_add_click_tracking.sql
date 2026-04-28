-- +goose Up
ALTER TABLE email_logs ADD COLUMN clicked_at TIMESTAMPTZ;

CREATE TABLE email_clicks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    log_id UUID NOT NULL REFERENCES email_logs(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    url_hash VARCHAR(16) NOT NULL,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET
);

CREATE INDEX idx_email_clicks_log_id ON email_clicks(log_id);
CREATE INDEX idx_email_clicks_url_hash ON email_clicks(url_hash);
CREATE INDEX idx_email_clicks_clicked_at ON email_clicks(clicked_at DESC);

-- +goose Down
DROP TABLE email_clicks;
ALTER TABLE email_logs DROP COLUMN clicked_at;

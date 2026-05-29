-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_email_logs_to_email_trgm
    ON email_logs USING gin (to_email gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_email_logs_subject_trgm
    ON email_logs USING gin (subject gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_email_logs_subject_trgm;
DROP INDEX IF EXISTS idx_email_logs_to_email_trgm;

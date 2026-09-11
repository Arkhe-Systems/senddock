-- +goose Up
-- Outgoing SMTP used to accept any certificate. Verification is on by default now, so a
-- project pointing at a relay with a self-signed or expired certificate needs an explicit
-- opt-out. Keeping it per project keeps the exception scoped and visible instead of
-- weakening every project on the instance.
ALTER TABLE projects ADD COLUMN smtp_allow_insecure_tls BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE projects DROP COLUMN smtp_allow_insecure_tls;

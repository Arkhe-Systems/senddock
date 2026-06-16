-- +goose Up
ALTER TABLE users
    ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified;

-- +goose Up
ALTER TABLE instance_settings ADD COLUMN license_key_encrypted TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE instance_settings DROP COLUMN license_key_encrypted;

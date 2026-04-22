-- +goose Up
ALTER TABLE campaigns ADD COLUMN variables JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE campaigns DROP COLUMN variables;

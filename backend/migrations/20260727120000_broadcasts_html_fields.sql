-- +goose Up
ALTER TABLE broadcasts ADD COLUMN html_fields JSONB NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE broadcasts DROP COLUMN html_fields;

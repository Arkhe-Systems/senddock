-- +goose Up
ALTER TABLE campaigns ADD COLUMN subject VARCHAR(998) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE campaigns DROP COLUMN subject;

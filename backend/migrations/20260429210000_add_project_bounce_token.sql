-- +goose Up
ALTER TABLE projects ADD COLUMN bounce_token UUID NOT NULL DEFAULT gen_random_uuid();

-- +goose Down
ALTER TABLE projects DROP COLUMN bounce_token;

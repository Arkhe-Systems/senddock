-- +goose Up
ALTER TABLE projects ADD COLUMN description TEXT;
ALTER TABLE projects ALTER COLUMN from_email DROP NOT NULL;
ALTER TABLE projects ALTER COLUMN from_name DROP NOT NULL;
ALTER TABLE projects ALTER COLUMN from_name DROP DEFAULT;

-- +goose Down
ALTER TABLE projects DROP COLUMN description;
ALTER TABLE projects ALTER COLUMN from_email SET NOT NULL;
ALTER TABLE projects ALTER COLUMN from_name SET NOT NULL;
ALTER TABLE projects ALTER COLUMN from_name SET DEFAULT 'SendDock';

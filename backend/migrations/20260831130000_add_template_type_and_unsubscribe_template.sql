-- +goose Up
ALTER TABLE templates ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'email';
ALTER TABLE templates ADD CONSTRAINT templates_type_check CHECK (type IN ('email', 'page'));

ALTER TABLE projects ADD COLUMN unsubscribe_template_id UUID REFERENCES templates(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE projects DROP COLUMN unsubscribe_template_id;
ALTER TABLE templates DROP CONSTRAINT templates_type_check;
ALTER TABLE templates DROP COLUMN type;

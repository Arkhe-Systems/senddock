-- +goose Up
CREATE TABLE subscriber_field_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key VARCHAR(64) NOT NULL,
    label VARCHAR(128) NOT NULL,
    field_type VARCHAR(20) NOT NULL,
    options JSONB,
    required BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, key)
);

CREATE INDEX idx_subscriber_field_definitions_project_id ON subscriber_field_definitions(project_id);

-- +goose Down
DROP TABLE subscriber_field_definitions;

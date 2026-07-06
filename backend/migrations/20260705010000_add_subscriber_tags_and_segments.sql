-- +goose Up
ALTER TABLE subscribers ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';
CREATE INDEX idx_subscribers_tags ON subscribers USING GIN (tags);

CREATE TABLE segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    predicate JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_segments_project_id ON segments(project_id);

-- +goose Down
DROP TABLE segments;
DROP INDEX idx_subscribers_tags;
ALTER TABLE subscribers DROP COLUMN tags;

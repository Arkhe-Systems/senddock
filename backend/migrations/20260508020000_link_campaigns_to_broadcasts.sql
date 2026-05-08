-- +goose Up
ALTER TABLE campaigns ADD COLUMN broadcast_id UUID REFERENCES broadcasts(id) ON DELETE SET NULL;
CREATE INDEX idx_campaigns_broadcast ON campaigns(broadcast_id) WHERE broadcast_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_campaigns_broadcast;
ALTER TABLE campaigns DROP COLUMN broadcast_id;

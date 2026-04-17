-- +goose Up
CREATE INDEX idx_subscribers_metadata ON subscribers USING GIN (metadata);

-- +goose Down
DROP INDEX idx_subscribers_metadata;

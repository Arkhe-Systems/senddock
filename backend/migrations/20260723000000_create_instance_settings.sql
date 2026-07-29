-- +goose Up
CREATE TABLE instance_settings (
    id BOOLEAN PRIMARY KEY DEFAULT true CHECK (id),
    public_url TEXT NOT NULL DEFAULT '',
    session_idle_timeout_minutes INTEGER NOT NULL DEFAULT 120 CHECK (session_idle_timeout_minutes > 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO instance_settings (id) VALUES (true);

-- +goose Down
DROP TABLE instance_settings;

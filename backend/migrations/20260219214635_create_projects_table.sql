-- +goose Up
  CREATE TABLE projects (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      name VARCHAR(255) NOT NULL,

      -- Sender domain
      from_name VARCHAR(255) NOT NULL DEFAULT 'SendDock',
      from_email VARCHAR(255) NOT NULL,

      -- SMTP Community
      smtp_host VARCHAR(255),
      smtp_port INT,
      smtp_user VARCHAR(255),
      smtp_password_encrypted TEXT,

      -- Webhook URL
      webhook_url TEXT,
      webhook_secret VARCHAR(255),

      -- Project config
      tracking_enabled BOOLEAN NOT NULL DEFAULT false,

      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE INDEX idx_projects_user_id ON projects(user_id);

-- +goose Down
  DROP TABLE projects;
-- +goose Up
-- +goose StatementBegin
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(16) NOT NULL CHECK (role IN ('owner', 'admin', 'developer', 'viewer')),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX workspace_members_user_idx ON workspace_members(user_id);

ALTER TABLE projects ADD COLUMN workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE;

INSERT INTO workspaces (id, name, created_by, created_at, updated_at)
SELECT gen_random_uuid(), 'My Workspace', u.id, NOW(), NOW()
FROM users u;

INSERT INTO workspace_members (workspace_id, user_id, role, created_at)
SELECT w.id, w.created_by, 'owner', NOW()
FROM workspaces w;

UPDATE projects p
SET workspace_id = w.id
FROM workspaces w
WHERE w.created_by = p.user_id
  AND p.workspace_id IS NULL;

ALTER TABLE projects ALTER COLUMN workspace_id SET NOT NULL;
CREATE INDEX projects_workspace_idx ON projects(workspace_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS projects_workspace_idx;
ALTER TABLE projects DROP COLUMN IF EXISTS workspace_id;
DROP INDEX IF EXISTS workspace_members_user_idx;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
-- +goose StatementEnd

-- +goose Up
ALTER TABLE projects
    ADD COLUMN bounce_imap_host VARCHAR(255),
    ADD COLUMN bounce_imap_port INTEGER,
    ADD COLUMN bounce_imap_user VARCHAR(255),
    ADD COLUMN bounce_imap_password_encrypted TEXT,
    ADD COLUMN bounce_imap_folder VARCHAR(64) NOT NULL DEFAULT 'INBOX',
    ADD COLUMN bounce_imap_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE projects
    DROP COLUMN bounce_imap_host,
    DROP COLUMN bounce_imap_port,
    DROP COLUMN bounce_imap_user,
    DROP COLUMN bounce_imap_password_encrypted,
    DROP COLUMN bounce_imap_folder,
    DROP COLUMN bounce_imap_enabled;

-- +goose Up
CREATE TABLE newsletters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_newsletters_project_id ON newsletters(project_id);

CREATE TABLE newsletter_subscriptions (
    newsletter_id UUID NOT NULL REFERENCES newsletters(id) ON DELETE CASCADE,
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    subscribed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unsubscribed_at TIMESTAMPTZ,
    PRIMARY KEY (newsletter_id, subscriber_id)
);

CREATE INDEX idx_newsletter_subscriptions_subscriber ON newsletter_subscriptions(subscriber_id);
CREATE INDEX idx_newsletter_subscriptions_project ON newsletter_subscriptions(project_id);

ALTER TABLE broadcasts ADD COLUMN newsletter_id UUID REFERENCES newsletters(id) ON DELETE SET NULL;
CREATE INDEX idx_broadcasts_newsletter ON broadcasts(newsletter_id) WHERE newsletter_id IS NOT NULL;

ALTER TABLE campaigns ADD COLUMN newsletter_id UUID REFERENCES newsletters(id) ON DELETE SET NULL;

ALTER TABLE email_logs ADD COLUMN newsletter_id UUID REFERENCES newsletters(id) ON DELETE SET NULL;
CREATE INDEX idx_email_logs_newsletter ON email_logs(newsletter_id) WHERE newsletter_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_email_logs_newsletter;
ALTER TABLE email_logs DROP COLUMN newsletter_id;
ALTER TABLE campaigns DROP COLUMN newsletter_id;
DROP INDEX IF EXISTS idx_broadcasts_newsletter;
ALTER TABLE broadcasts DROP COLUMN newsletter_id;
DROP INDEX IF EXISTS idx_newsletter_subscriptions_project;
DROP INDEX IF EXISTS idx_newsletter_subscriptions_subscriber;
DROP TABLE newsletter_subscriptions;
DROP INDEX IF EXISTS idx_newsletters_project_id;
DROP TABLE newsletters;

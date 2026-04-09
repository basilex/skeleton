-- Migration: Create notification_templates table
-- Email/SMS/Push notification templates

CREATE TABLE notification_templates (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name        TEXT NOT NULL UNIQUE,
    channel     TEXT NOT NULL CHECK (channel IN ('email', 'sms', 'push')),
    subject     TEXT,
    body        TEXT NOT NULL,
    variables   JSONB DEFAULT '{}',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_templates_name ON notification_templates(name);
CREATE INDEX idx_notification_templates_channel ON notification_templates(channel) WHERE is_active = TRUE;

COMMENT ON TABLE notification_templates IS 'Notification templates with UUID v7';
COMMENT ON COLUMN notification_templates.id IS 'UUID v7 primary key';
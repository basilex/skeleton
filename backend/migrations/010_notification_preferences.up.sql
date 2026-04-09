-- Migration: Create notification_preferences table
-- User notification settings

CREATE TABLE notification_preferences (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel           TEXT NOT NULL CHECK (channel IN ('email', 'sms', 'push')),
    is_enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    preferences       JSONB DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, channel)
);

CREATE INDEX idx_notification_preferences_user ON notification_preferences(user_id);

COMMENT ON TABLE notification_preferences IS 'User notification preferences';
COMMENT ON COLUMN notification_preferences.id IS 'UUID v7 primary key';
-- Migration: Create notifications table
-- Notification queue with UUID and JSONB metadata

CREATE TABLE notifications (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id       UUID REFERENCES users(id) ON DELETE CASCADE,
    email         TEXT,
    phone         TEXT,
    device_token  TEXT,
    channel       TEXT NOT NULL,
    subject       TEXT,
    content       TEXT NOT NULL,
    html_content  TEXT,
    status        TEXT NOT NULL DEFAULT 'pending',
    priority      TEXT NOT NULL DEFAULT 'normal',
    scheduled_at  TIMESTAMPTZ,
    sent_at       TIMESTAMPTZ,
    delivered_at  TIMESTAMPTZ,
    failed_at     TIMESTAMPTZ,
    failure_reason TEXT,
    attempts      INTEGER NOT NULL DEFAULT 0,
    max_attempts  INTEGER NOT NULL DEFAULT 5,
    metadata      JSONB DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_priority ON notifications(priority);
CREATE INDEX idx_notifications_scheduled ON notifications(scheduled_at) WHERE status = 'pending';
CREATE INDEX idx_notifications_created ON notifications(created_at);
CREATE INDEX idx_notifications_metadata_gin ON notifications USING GIN (metadata);

-- Partial indexes for performance
CREATE INDEX idx_notifications_pending ON notifications(id) WHERE status = 'pending';

COMMENT ON TABLE notifications IS 'Notification queue with UUID v7 primary key and JSONB metadata';
COMMENT ON COLUMN notifications.id IS 'UUID v7 primary key - time-sortable';
COMMENT ON COLUMN notifications.user_id IS 'FK to users.id (UUID)';
COMMENT ON COLUMN notifications.metadata IS 'JSONB field for notification metadata';
-- Sessions table for user session management
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    device_type TEXT,
    os TEXT,
    browser TEXT,
    device_name TEXT,
    user_agent TEXT,
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    last_activity TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT
);

-- User preferences table
CREATE TABLE user_preferences (
    id TEXT PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    theme TEXT NOT NULL DEFAULT 'auto',
    language TEXT NOT NULL DEFAULT 'en',
    date_format TEXT NOT NULL DEFAULT 'yyyy-mm-dd',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    email_notifications BOOLEAN NOT NULL DEFAULT true,
    sms_notifications BOOLEAN NOT NULL DEFAULT false,
    push_notifications BOOLEAN NOT NULL DEFAULT true,
    in_app_notifications BOOLEAN NOT NULL DEFAULT true,
    marketing_emails BOOLEAN NOT NULL DEFAULT false,
    weekly_digest BOOLEAN NOT NULL DEFAULT true,
    quiet_hours_start INTEGER,
    quiet_hours_end INTEGER,
    notifications_timezone TEXT DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for sessions
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_last_activity ON sessions(last_activity);

-- Index for user_preferences
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Function to automatically clean up expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM sessions WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;
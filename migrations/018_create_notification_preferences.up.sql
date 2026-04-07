CREATE TABLE notification_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT UNIQUE NOT NULL,
    preferences TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_notification_preferences_user ON notification_preferences(user_id);
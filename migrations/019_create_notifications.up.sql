CREATE TABLE notifications (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    email TEXT,
    phone TEXT,
    device_token TEXT,
    channel TEXT NOT NULL,
    subject TEXT,
    content TEXT NOT NULL,
    html_content TEXT,
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    scheduled_at TEXT,
    sent_at TEXT,
    delivered_at TEXT,
    failed_at TEXT,
    failure_reason TEXT,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 5,
    metadata TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_scheduled ON notifications(scheduled_at);
CREATE INDEX idx_notifications_created ON notifications(created_at);
CREATE INDEX idx_notifications_priority ON notifications(priority);
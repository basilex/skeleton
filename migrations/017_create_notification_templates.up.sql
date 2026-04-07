CREATE TABLE notification_templates (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    channel TEXT NOT NULL,
    subject TEXT,
    body TEXT NOT NULL,
    html_body TEXT,
    variables TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX idx_notification_templates_name ON notification_templates(name);
CREATE INDEX idx_notification_templates_channel ON notification_templates(channel);
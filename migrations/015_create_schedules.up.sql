CREATE TABLE task_schedules (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    task_type TEXT NOT NULL,
    payload TEXT NOT NULL,
    cron TEXT NOT NULL,
    timezone TEXT DEFAULT 'UTC',
    last_run_at TEXT,
    next_run_at TEXT,
    is_active INTEGER DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_schedules_next_run ON task_schedules(next_run_at);
CREATE INDEX idx_schedules_active ON task_schedules(is_active);
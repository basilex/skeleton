CREATE TABLE dead_letters (
    id TEXT PRIMARY KEY,
    original_task_id TEXT NOT NULL,
    original_task TEXT NOT NULL,
    failed_at TEXT NOT NULL,
    reason TEXT NOT NULL,
    reviewed INTEGER DEFAULT 0,
    reviewed_at TEXT,
    reviewed_by TEXT,
    action TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_dead_letters_reviewed ON dead_letters(reviewed);
CREATE INDEX idx_dead_letters_failed_at ON dead_letters(failed_at);
-- Migration: Create dead_letters table
-- Failed tasks queue

CREATE TABLE dead_letters (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    task_type     TEXT NOT NULL,
    payload       JSONB NOT NULL DEFAULT '{}',
    error_message TEXT NOT NULL,
    attempts      INTEGER NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dead_letters_created ON dead_letters(created_at);

COMMENT ON TABLE dead_letters IS 'Failed tasks for manual review';
COMMENT ON COLUMN dead_letters.id IS 'UUID v7 primary key';
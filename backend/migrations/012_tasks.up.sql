-- Migration: Create tasks table
-- Background task processing with UUID

CREATE TABLE tasks (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    task_type      TEXT NOT NULL,
    payload        JSONB NOT NULL DEFAULT '{}',
    status         TEXT NOT NULL DEFAULT 'pending',
    priority       INTEGER NOT NULL DEFAULT 5,
    attempts       INTEGER NOT NULL DEFAULT 0,
    max_attempts   INTEGER NOT NULL DEFAULT 5,
    error_message  TEXT,
    scheduled_at   TIMESTAMPTZ,
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- Helper Functions for Tasks
-- ============================================================================

-- Function to mark stalled tasks as failed
CREATE OR REPLACE FUNCTION mark_stalled_tasks_failed(stalled_after INTERVAL DEFAULT INTERVAL '30 minutes')
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    UPDATE tasks
    SET status = 'failed',
        error_message = 'Task stalled and marked as failed',
        updated_at = NOW()
    WHERE status = 'running'
      AND updated_at < NOW() - stalled_after;
    
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION mark_stalled_tasks_failed(INTERVAL) IS 
    'Marks tasks that have been running too long as failed and returns the count.';

-- ============================================================================
-- Indexes for common queries
-- ============================================================================
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_type ON tasks(task_type);
CREATE INDEX idx_tasks_scheduled ON tasks(scheduled_at) WHERE status = 'pending';
CREATE INDEX idx_tasks_payload_gin ON tasks USING GIN (payload);

COMMENT ON TABLE tasks IS 'Background tasks with UUID v7 and JSONB payload';
COMMENT ON COLUMN tasks.id IS 'UUID v7 primary key';
COMMENT ON COLUMN tasks.payload IS 'JSONB task payload - queryable with GIN index';
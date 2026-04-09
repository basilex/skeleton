-- Migration: Create task_schedules table
-- Recurring task scheduling

CREATE TABLE task_schedules (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    task_type      TEXT NOT NULL,
    payload        JSONB NOT NULL DEFAULT '{}',
    cron_schedule  TEXT NOT NULL,
    last_run_at    TIMESTAMPTZ,
    next_run_at    TIMESTAMPTZ NOT NULL,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_schedules_next_run ON task_schedules(next_run_at) WHERE is_active = TRUE;

COMMENT ON TABLE task_schedules IS 'Scheduled recurring tasks';
COMMENT ON COLUMN task_schedules.id IS 'UUID v7 primary key';
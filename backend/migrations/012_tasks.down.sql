-- Migration: Drop tasks table

-- Drop helper function
DROP FUNCTION IF EXISTS mark_stalled_tasks_failed(INTERVAL);

-- Drop table
DROP TABLE IF EXISTS tasks CASCADE;
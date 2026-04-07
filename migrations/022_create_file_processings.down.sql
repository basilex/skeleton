-- Drop file_processings table

DROP INDEX IF EXISTS idx_processings_created_at;
DROP INDEX IF EXISTS idx_processings_operation;
DROP INDEX IF EXISTS idx_processings_status;
DROP INDEX IF EXISTS idx_processings_file_id;

DROP TABLE IF EXISTS file_processings;
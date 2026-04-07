-- Drop file_uploads table

DROP INDEX IF EXISTS idx_uploads_created_at;
DROP INDEX IF EXISTS idx_uploads_expires_at;
DROP INDEX IF EXISTS idx_uploads_status;
DROP INDEX IF EXISTS idx_uploads_file_id;

DROP TABLE IF EXISTS file_uploads;
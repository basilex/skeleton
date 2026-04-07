-- Drop files table

DROP INDEX IF EXISTS idx_files_created_at;
DROP INDEX IF EXISTS idx_files_expires_at;
DROP INDEX IF EXISTS idx_files_uploaded_at;
DROP INDEX IF EXISTS idx_files_access_level;
DROP INDEX IF EXISTS idx_files_storage_provider;
DROP INDEX IF EXISTS idx_files_mime_type;
DROP INDEX IF EXISTS idx_files_owner;

DROP TABLE IF EXISTS files;
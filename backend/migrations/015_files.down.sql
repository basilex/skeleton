-- Migration: Drop files table

-- Drop helper function
DROP FUNCTION IF EXISTS clean_expired_files();

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS file_storage_stats CASCADE;

-- Drop table
DROP TABLE IF EXISTS files CASCADE;
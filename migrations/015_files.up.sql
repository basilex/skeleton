-- Migration: Create files table
-- File metadata with UUID and JSONB

CREATE TABLE files (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    owner_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    filename         TEXT NOT NULL,
    stored_name      TEXT NOT NULL,
    mime_type        TEXT NOT NULL,
    size             INTEGER NOT NULL,
    path             TEXT NOT NULL,
    storage_provider TEXT NOT NULL CHECK (storage_provider IN ('local', 's3', 'gcs')),
    checksum         TEXT NOT NULL,
    metadata         JSONB NOT NULL DEFAULT '{}',
    access_level     TEXT NOT NULL CHECK (access_level IN ('public', 'private', 'restricted')),
    uploaded_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMPTZ,
    processed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Generated column for file extension
    file_extension VARCHAR(10) GENERATED ALWAYS AS (LOWER(SPLIT_PART(filename, '.', ARRAY_LENGTH(STRING_TO_ARRAY(filename, '.'), 1)))) STORED
);

-- ============================================================================
-- Helper Functions for Files
-- ============================================================================

-- Function to clean expired files
CREATE OR REPLACE FUNCTION clean_expired_files()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM files 
    WHERE expires_at IS NOT NULL AND expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION clean_expired_files() IS 
    'Deletes expired files and returns the count of deleted rows.';

-- ============================================================================
-- Indexes
-- ============================================================================
CREATE INDEX idx_files_owner ON files(owner_id);
CREATE INDEX idx_files_mime_type ON files(mime_type);
CREATE INDEX idx_files_storage_provider ON files(storage_provider);
CREATE INDEX idx_files_access_level ON files(access_level);
CREATE INDEX idx_files_uploaded_at ON files(uploaded_at);
CREATE INDEX idx_files_expires_at ON files(expires_at);
CREATE INDEX idx_files_created_at ON files(created_at);

-- GIN index for JSONB metadata
CREATE INDEX idx_files_metadata_gin ON files USING GIN (metadata);

-- Materialized view for file statistics
CREATE MATERIALIZED VIEW file_storage_stats AS
SELECT 
    storage_provider,
    COUNT(*) AS total_files,
    SUM(size) AS total_size,
    AVG(size) AS avg_file_size,
    COUNT(*) FILTER (WHERE access_level = 'public') AS public_files,
    COUNT(*) FILTER (WHERE access_level = 'private') AS private_files
FROM files
GROUP BY storage_provider;

CREATE UNIQUE INDEX idx_file_storage_stats_provider ON file_storage_stats (storage_provider);

COMMENT ON TABLE files IS 'File metadata with UUID v7 primary key and JSONB metadata';
COMMENT ON COLUMN files.id IS 'UUID v7 primary key - time-sortable';
COMMENT ON COLUMN files.owner_id IS 'FK to users.id (UUID)';
COMMENT ON COLUMN files.metadata IS 'JSONB field for extensible metadata - queryable with ->> and @> operators';
-- Migration: Create file_uploads table
-- Tracks file upload processes

CREATE TABLE file_uploads (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    file_id     UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    chunk_hash TEXT NOT NULL,
    
    UNIQUE(file_id, chunk_index)
);

CREATE INDEX idx_file_uploads_file ON file_uploads(file_id);

COMMENT ON TABLE file_uploads IS 'File chunk uploads for large files';
COMMENT ON COLUMN file_uploads.id IS 'UUID v7 primary key';
COMMENT ON COLUMN file_uploads.file_id IS 'FK to files.id (UUID)';
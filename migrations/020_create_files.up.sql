-- Create files table
-- This table stores file metadata and access information.

CREATE TABLE files (
    id TEXT PRIMARY KEY NOT NULL,
    owner_id TEXT,
    filename TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    path TEXT NOT NULL,
    storage_provider TEXT NOT NULL CHECK(storage_provider IN ('local', 's3', 'gcs')),
    checksum TEXT NOT NULL,
    metadata TEXT, -- JSON object: {width, height, duration, pages, thumbnail, original_id, custom}
    access_level TEXT NOT NULL CHECK(access_level IN ('public', 'private', 'restricted')),
    uploaded_at TEXT NOT NULL,
    expires_at TEXT,
    processed_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Indexes for common queries
CREATE INDEX idx_files_owner ON files(owner_id);
CREATE INDEX idx_files_mime_type ON files(mime_type);
CREATE INDEX idx_files_storage_provider ON files(storage_provider);
CREATE INDEX idx_files_access_level ON files(access_level);
CREATE INDEX idx_files_uploaded_at ON files(uploaded_at);
CREATE INDEX idx_files_expires_at ON files(expires_at);
CREATE INDEX idx_files_created_at ON files(created_at);
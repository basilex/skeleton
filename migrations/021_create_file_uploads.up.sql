-- Create file_uploads table
-- This table tracks presigned upload URLs for large file uploads (>5MB).

CREATE TABLE file_uploads (
    id TEXT PRIMARY KEY NOT NULL,
    file_id TEXT NOT NULL,
    upload_url TEXT NOT NULL,
    fields TEXT, -- JSON object with additional form fields for upload
    status TEXT NOT NULL CHECK(status IN ('pending', 'completed', 'failed', 'expired')),
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Indexes for common queries
CREATE INDEX idx_uploads_file_id ON file_uploads(file_id);
CREATE INDEX idx_uploads_status ON file_uploads(status);
CREATE INDEX idx_uploads_expires_at ON file_uploads(expires_at);
CREATE INDEX idx_uploads_created_at ON file_uploads(created_at);
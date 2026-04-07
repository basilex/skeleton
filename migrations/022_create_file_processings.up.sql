-- Create file_processings table
-- This table tracks file processing operations (resize, crop, compress, etc.).

CREATE TABLE file_processings (
    id TEXT PRIMARY KEY NOT NULL,
    file_id TEXT NOT NULL,
    operation TEXT NOT NULL CHECK(operation IN ('resize', 'crop', 'compress', 'convert', 'thumbnail', 'watermark')),
    options TEXT, -- JSON object: {width, height, x, y, quality, format, watermark, custom}
    status TEXT NOT NULL CHECK(status IN ('pending', 'running', 'completed', 'failed')),
    result_file_id TEXT,
    error TEXT,
    started_at TEXT,
    completed_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (result_file_id) REFERENCES files(id) ON DELETE SET NULL
);

-- Indexes for common queries
CREATE INDEX idx_processings_file_id ON file_processings(file_id);
CREATE INDEX idx_processings_status ON file_processings(status);
CREATE INDEX idx_processings_operation ON file_processings(operation);
CREATE INDEX idx_processings_created_at ON file_processings(created_at);
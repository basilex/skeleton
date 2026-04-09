-- Migration: Create file_processings table
-- File processing pipeline jobs

CREATE TABLE file_processings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    file_id         UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    operation       TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    result_file_id  UUID REFERENCES files(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX idx_file_processings_file ON file_processings(file_id);
CREATE INDEX idx_file_processings_status ON file_processings(status);

COMMENT ON TABLE file_processings IS 'File processing jobs';
COMMENT ON COLUMN file_processings.id IS 'UUID v7 primary key';
COMMENT ON COLUMN file_processings.file_id IS 'FK to files.id (UUID)';
COMMENT ON COLUMN file_processings.result_file_id IS 'FK to files.id for processed result (UUID)';
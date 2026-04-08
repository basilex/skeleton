-- Migration: Create audit_records table
-- Stores audit trail with UUID primary key

CREATE TABLE audit_records (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    actor_id    UUID NOT NULL,
    actor_type  TEXT NOT NULL,
    action      TEXT NOT NULL,
    resource    TEXT NOT NULL,
    resource_id UUID,
    metadata    JSONB DEFAULT '{}',
    ip          TEXT,
    user_agent  TEXT,
    status      INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_audit_actor_id ON audit_records(actor_id);
CREATE INDEX idx_audit_resource ON audit_records(resource);
CREATE INDEX idx_audit_created_at ON audit_records(created_at);
CREATE INDEX idx_audit_metadata_gin ON audit_records USING GIN (metadata);

COMMENT ON TABLE audit_records IS 'Audit trail with UUID v7 primary key and JSONB metadata';
COMMENT ON COLUMN audit_records.id IS 'UUID v7 primary key - time-sortable';
COMMENT ON COLUMN audit_records.metadata IS 'JSONB field for flexible audit metadata';
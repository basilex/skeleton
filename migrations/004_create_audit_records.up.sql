CREATE TABLE audit_records (
    id TEXT PRIMARY KEY,
    actor_id TEXT NOT NULL,
    actor_type TEXT NOT NULL,
    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    resource_id TEXT,
    metadata TEXT,
    ip TEXT,
    user_agent TEXT,
    status INTEGER NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_audit_actor_id ON audit_records(actor_id);
CREATE INDEX idx_audit_resource ON audit_records(resource);
CREATE INDEX idx_audit_created_at ON audit_records(created_at);
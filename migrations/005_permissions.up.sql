-- Migration: Create permissions table

CREATE TABLE permissions (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permissions_name ON permissions(name);

COMMENT ON TABLE permissions IS 'Permissions for role-based access control';
COMMENT ON COLUMN permissions.id IS 'UUID v7 primary key';
-- Migration: Create roles table
-- Primary key uses UUID v7 (time-sortable)

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_roles_name ON roles(name);

COMMENT ON TABLE roles IS 'User roles (admin, user, moderator, etc.) with UUID v7 primary key';
COMMENT ON COLUMN roles.id IS 'UUID v7 primary key - time-sortable';
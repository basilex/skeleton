-- Migration: Create users table
-- Primary key uses UUID v7 (time-sortable)

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(email) WHERE is_active = TRUE;

-- Full-text search index on email (optional, for email search)
CREATE INDEX idx_users_email_gin ON users USING GIN (to_tsvector('simple', email));

COMMENT ON TABLE users IS 'Application users with UUID v7 primary key';
COMMENT ON COLUMN users.id IS 'UUID v7 primary key - time-sortable';
COMMENT ON COLUMN users.is_active IS 'Whether user account is active (deleted users set to false)';
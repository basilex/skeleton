-- Migration: Create user_roles junction table
-- Links users to roles with UUID foreign keys

CREATE TABLE user_roles (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (user_id, role_id)
);

-- Indexes
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

COMMENT ON TABLE user_roles IS 'Junction table linking users to their roles';
COMMENT ON COLUMN user_roles.user_id IS 'FK to users.id (UUID)';
COMMENT ON COLUMN user_roles.role_id IS 'FK to roles.id (UUID)';
CREATE TABLE user_roles (
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TEXT NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE refresh_tokens (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);

CREATE TABLE roles (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);

CREATE TABLE permissions (
    id       TEXT PRIMARY KEY,
    name     TEXT NOT NULL UNIQUE,
    resource TEXT NOT NULL,
    action   TEXT NOT NULL
);

CREATE TABLE role_permissions (
    role_id       TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

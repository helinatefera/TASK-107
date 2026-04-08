CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE role_permissions (
    role          TEXT NOT NULL CHECK (role IN ('guest','user','merchant','administrator')),
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_id)
);

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('guest','user','merchant','administrator')),
    org_id        UUID,
    locked_until  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_permissions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, permission_id)
);

CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id       TEXT NOT NULL,
    token_hash      TEXT NOT NULL UNIQUE,
    idle_expires_at TIMESTAMPTZ NOT NULL,
    abs_expires_at  TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_user_device ON sessions(user_id, device_id);
CREATE INDEX idx_sessions_expiry ON sessions(idle_expires_at);

CREATE TABLE recovery_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE login_attempts (
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    success      BOOLEAN NOT NULL,
    ip_address   INET,
    device_id    TEXT,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_login_attempts_user_time ON login_attempts(user_id, attempted_at DESC);

-- Seed default permissions
INSERT INTO permissions (id, name, description) VALUES
    (gen_random_uuid(), 'user.read', 'View user profiles'),
    (gen_random_uuid(), 'user.manage', 'Manage users and roles'),
    (gen_random_uuid(), 'org.read', 'View organizations'),
    (gen_random_uuid(), 'org.manage', 'Manage organizations'),
    (gen_random_uuid(), 'warehouse.read', 'View warehouses'),
    (gen_random_uuid(), 'warehouse.manage', 'Manage warehouses'),
    (gen_random_uuid(), 'item.read', 'View items and categories'),
    (gen_random_uuid(), 'item.manage', 'Manage items and categories'),
    (gen_random_uuid(), 'supplier.read', 'View suppliers and carriers'),
    (gen_random_uuid(), 'supplier.manage', 'Manage suppliers and carriers'),
    (gen_random_uuid(), 'station.read', 'View stations and devices'),
    (gen_random_uuid(), 'station.manage', 'Manage stations and devices'),
    (gen_random_uuid(), 'pricing.read', 'View pricing templates'),
    (gen_random_uuid(), 'pricing.manage', 'Manage pricing templates'),
    (gen_random_uuid(), 'order.read', 'View orders'),
    (gen_random_uuid(), 'order.create', 'Create orders'),
    (gen_random_uuid(), 'content.read', 'View content modules'),
    (gen_random_uuid(), 'content.manage', 'Manage content modules'),
    (gen_random_uuid(), 'notification.read', 'View notifications'),
    (gen_random_uuid(), 'notification.manage', 'Manage notification templates'),
    (gen_random_uuid(), 'admin.config', 'Manage global configuration'),
    (gen_random_uuid(), 'admin.audit', 'View audit logs'),
    (gen_random_uuid(), 'admin.metrics', 'View and export metrics');

-- Seed role-permission mappings
INSERT INTO role_permissions (role, permission_id)
SELECT 'administrator', id FROM permissions;

INSERT INTO role_permissions (role, permission_id)
SELECT 'merchant', id FROM permissions WHERE name IN (
    'org.read', 'warehouse.read', 'warehouse.manage', 'item.read', 'item.manage',
    'supplier.read', 'supplier.manage', 'station.read', 'station.manage',
    'pricing.read', 'pricing.manage', 'order.read', 'order.create',
    'content.read', 'content.manage', 'notification.read'
);

INSERT INTO role_permissions (role, permission_id)
SELECT 'user', id FROM permissions WHERE name IN (
    'org.read', 'item.read', 'station.read', 'pricing.read',
    'order.read', 'order.create', 'content.read', 'notification.read'
);

INSERT INTO role_permissions (role, permission_id)
SELECT 'guest', id FROM permissions WHERE name IN (
    'content.read'
);

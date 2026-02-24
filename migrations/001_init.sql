-- Master-Slave Server: Initial Schema
-- Run against PostgreSQL 16+

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- USERS
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- ============================================================
-- APP REGISTRY (Slave apps)
-- ============================================================
CREATE TABLE IF NOT EXISTS app_registry (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name         VARCHAR(255) NOT NULL,
    package_id       VARCHAR(255) NOT NULL UNIQUE,
    deep_link_scheme VARCHAR(255) NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- USER ↔ APP PERMISSIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS user_app_permissions (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_id  UUID NOT NULL REFERENCES app_registry(id) ON DELETE CASCADE,
    UNIQUE(user_id, app_id)
);

CREATE INDEX idx_user_app_permissions_user_id ON user_app_permissions(user_id);
CREATE INDEX idx_user_app_permissions_app_id  ON user_app_permissions(app_id);

-- ============================================================
-- ONE-TIME CODES (OTC Handshake)
-- ============================================================
CREATE TABLE IF NOT EXISTS one_time_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_id     UUID        NOT NULL REFERENCES app_registry(id) ON DELETE CASCADE,
    code       VARCHAR(32) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    claimed    BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_one_time_codes_code       ON one_time_codes(code);
CREATE INDEX idx_one_time_codes_expires_at ON one_time_codes(expires_at);

-- ============================================================
-- SEED DATA (for development/testing)
-- ============================================================

-- Test user: admin@cachatto.click / password123
-- bcrypt hash of "password123"
INSERT INTO users (id, email, password_hash) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@cachatto.click', '$2a$10$uddi6eclOX5sEPYRYGQu/ud4fGBCpa.i5/weYHrcasPLHpeVfsQwi')
ON CONFLICT (email) DO NOTHING;

-- Sample slave apps
INSERT INTO app_registry (id, app_name, package_id, deep_link_scheme) VALUES
    ('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Slave App One',   'com.cachatto.slave1', 'slaveapp1://'),
    ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'Slave App Two',   'com.cachatto.slave2', 'slaveapp2://')
ON CONFLICT (package_id) DO NOTHING;

-- Grant admin access to both slave apps
INSERT INTO user_app_permissions (user_id, app_id) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33')
ON CONFLICT (user_id, app_id) DO NOTHING;

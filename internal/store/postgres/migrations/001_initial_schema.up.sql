-- 001_initial_schema.up.sql
-- Optimized Version following OpenTrusty Identity & Authorization Model principles.

-- -----------------------------------------------------------------------------
-- Core Principles (Codebase-Wide)
-- 1. No tenant represents the platform
-- 2. Platform authorization is expressed only via scoped roles
-- 3. Tenant context must never be elevated to platform context
--
-- Anti-Patterns (FORBIDDEN):
-- - Magic tenant IDs (e.g., "default", "system", "platform")
-- - Empty/NULL tenant_id implying platform privileges
-- - Hardcoded role checks (use permission checks)
-- -----------------------------------------------------------------------------

-- -----------------------------------------------------------------------------
-- 1. Scoped RBAC Tables (Foundation)
-- -----------------------------------------------------------------------------

-- Permissions Table (Normalization of actions)
CREATE TABLE IF NOT EXISTS rbac_permissions (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Scoped Roles Table
-- scope IN ('platform', 'tenant', 'client')
CREATE TABLE IF NOT EXISTS rbac_roles (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('platform', 'tenant', 'client')),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, scope)
);

-- Role-Permission Mapping
CREATE TABLE IF NOT EXISTS rbac_role_permissions (
    role_id VARCHAR(255) NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
    permission_id VARCHAR(255) NOT NULL REFERENCES rbac_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- -----------------------------------------------------------------------------
-- 2. Core Identity Tables
-- -----------------------------------------------------------------------------

-- Local helper for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Tenants Table
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Seed Sample Tenant (Non-privileged, for initial user registration)
INSERT INTO tenants (id, name) VALUES ('sample', 'Sample Tenant') ON CONFLICT (id) DO NOTHING;

-- Users Table (Identities)
-- NOTE: tenant_id is NOT NULL. All users belong to a tenant.
-- Platform Admin privileges are derived from rbac_assignments, NOT from tenant_id.
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Profile information
    given_name VARCHAR(255),
    family_name VARCHAR(255),
    full_name VARCHAR(255),
    nickname VARCHAR(255),
    picture TEXT,
    locale VARCHAR(10),
    timezone VARCHAR(50),
    
    -- Lockout management
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Ensure email is unique per tenant
    UNIQUE(tenant_id, email)
);

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Credentials Table
CREATE TABLE IF NOT EXISTS credentials (
    user_id VARCHAR(255) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Sessions Table
-- NOTE: tenant_id is NOT NULL.
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 3. Scoped RBAC Assignments
-- -----------------------------------------------------------------------------

-- Universal RBAC Assignments Table
-- This table implements "Platform authorization is expressed only via scoped roles"
-- scope_context_id = NULL is ONLY valid for scope = 'platform'
CREATE TABLE IF NOT EXISTS rbac_assignments (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id VARCHAR(255) NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('platform', 'tenant', 'client')),
    scope_context_id VARCHAR(255), -- NULL for platform, tenant_id for tenant, client_id for client
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by VARCHAR(255) REFERENCES users(id),
    
    UNIQUE(user_id, role_id, scope, scope_context_id),
    -- Constraint: scope_context_id must be NULL for platform scope
    CHECK ((scope = 'platform' AND scope_context_id IS NULL) OR (scope != 'platform' AND scope_context_id IS NOT NULL))
);

-- Keep Projects for grouping resources
CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 4. OAuth2 & OIDC Support
-- -----------------------------------------------------------------------------

-- OAuth2 Clients
CREATE TABLE IF NOT EXISTS oauth2_clients (
    id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    client_secret_hash TEXT NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    client_uri TEXT,
    logo_uri TEXT,
    redirect_uris JSONB NOT NULL DEFAULT '[]'::jsonb,
    allowed_scopes JSONB NOT NULL DEFAULT '["openid"]'::jsonb,
    grant_types JSONB NOT NULL DEFAULT '["authorization_code"]'::jsonb,
    response_types JSONB NOT NULL DEFAULT '["code"]'::jsonb,
    token_endpoint_auth_method VARCHAR(50) DEFAULT 'client_secret_basic',
    access_token_lifetime INTEGER DEFAULT 3600,
    refresh_token_lifetime INTEGER DEFAULT 2592000,
    id_token_lifetime INTEGER DEFAULT 3600,
    owner_id VARCHAR(255) REFERENCES users(id) ON DELETE SET NULL,
    is_trusted BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Authorization Codes
CREATE TABLE IF NOT EXISTS authorization_codes (
    id VARCHAR(255) PRIMARY KEY,
    code VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    scope TEXT NOT NULL,
    state TEXT,
    nonce TEXT,
    code_challenge TEXT,
    code_challenge_method VARCHAR(10),
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    is_used BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Access Tokens
CREATE TABLE IF NOT EXISTS access_tokens (
    id VARCHAR(255) PRIMARY KEY,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    is_revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Refresh Tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(255) PRIMARY KEY,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    access_token_id VARCHAR(255) REFERENCES access_tokens(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    is_revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- OpenID Keys
CREATE TABLE IF NOT EXISTS openid_keys (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    public_key TEXT NOT NULL,
    private_key_encrypted BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

-- -----------------------------------------------------------------------------
-- 5. Seeding & Utilities
-- -----------------------------------------------------------------------------

-- Triggers for updated_at
DROP TRIGGER IF EXISTS update_rbac_roles_updated_at ON rbac_roles;
CREATE TRIGGER update_rbac_roles_updated_at BEFORE UPDATE ON rbac_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_oauth2_clients_updated_at ON oauth2_clients;
CREATE TRIGGER update_oauth2_clients_updated_at BEFORE UPDATE ON oauth2_clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Seed Permissions
INSERT INTO rbac_permissions (id, name, description) VALUES
    ('p:tenant:create', 'tenant:create', 'Create new tenants'),
    ('p:tenant:delete', 'tenant:delete', 'Delete tenants'),
    ('p:tenant:list', 'tenant:list', 'List all tenants'),
    ('p:user:provision', 'user:provision', 'Provision users in a tenant'),
    ('p:user:manage', 'user:manage', 'Manage users in a tenant'),
    ('p:client:register', 'client:register', 'Register OAuth2 clients')
ON CONFLICT (id) DO NOTHING;

-- Seed Scoped Roles
INSERT INTO rbac_roles (id, name, scope, description) VALUES
    ('role:platform:admin', 'platform_admin', 'platform', 'Platform-wide administrator'),
    ('role:tenant:admin', 'tenant_admin', 'tenant', 'Administrator for a specific tenant'),
    ('role:tenant:member', 'member', 'tenant', 'Regular member of a tenant')
ON CONFLICT (id) DO NOTHING;

-- Map Permissions to Roles
-- Platform Admin: All
INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT 'role:platform:admin', id FROM rbac_permissions ON CONFLICT DO NOTHING;

-- Tenant Admin: Tenant-level management
INSERT INTO rbac_role_permissions (role_id, permission_id) VALUES
    ('role:tenant:admin', 'p:user:provision'),
    ('role:tenant:admin', 'p:user:manage'),
    ('role:tenant:admin', 'p:client:register')
ON CONFLICT DO NOTHING;

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
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Scoped Roles Table
-- scope IN ('platform', 'tenant', 'client')
CREATE TABLE IF NOT EXISTS rbac_roles (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('platform', 'tenant', 'client')),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, scope)
);

-- Role-Permission Mapping
CREATE TABLE IF NOT EXISTS rbac_role_permissions (
    role_id UUID NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES rbac_permissions(id) ON DELETE CASCADE,
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
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Users Table (Identities)
-- NOTE: tenant_id is NULLABLE for Platform Admins. 
-- Tenant Users MUST have a tenant_id.
-- Platform Admin privileges are derived from rbac_assignments, NOT from tenant_id.
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT,
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
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Sessions Table
-- NOTE: tenant_id is NULLABLE for Platform sessions.
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
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
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('platform', 'tenant', 'client')),
    scope_context_id UUID, -- NULL for platform, tenant_id for tenant, client_id for client
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by UUID REFERENCES users(id),
    
    UNIQUE(user_id, role_id, scope, scope_context_id),
    -- Constraint: scope_context_id must be NULL for platform scope
    CHECK ((scope = 'platform' AND scope_context_id IS NULL) OR (scope != 'platform' AND scope_context_id IS NOT NULL))
);

-- Keep Projects for grouping resources
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 4. OAuth2 & OIDC Support
-- -----------------------------------------------------------------------------

-- OAuth2 Clients
CREATE TABLE IF NOT EXISTS oauth2_clients (
    id UUID PRIMARY KEY,
    client_id UUID UNIQUE NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
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
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    is_trusted BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Authorization Codes
CREATE TABLE IF NOT EXISTS authorization_codes (
    id UUID PRIMARY KEY,
    code VARCHAR(255) UNIQUE NOT NULL,
    client_id UUID NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
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
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    client_id UUID NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    is_revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_access_tokens_tenant ON access_tokens(tenant_id);

-- Refresh Tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    access_token_id UUID REFERENCES access_tokens(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    is_revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_tenant ON refresh_tokens(tenant_id);

-- OpenID Keys
CREATE TABLE IF NOT EXISTS openid_keys (
    id UUID PRIMARY KEY,
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
    ('10000000-0000-0000-0000-000000000001', 'platform:manage_tenants', 'Create, update, and delete tenants'),
    ('10000000-0000-0000-0000-000000000002', 'platform:manage_admins', 'Manage platform administrators'),
    ('10000000-0000-0000-0000-000000000003', 'platform:view_audit', 'View platform audit logs'),
    ('10000000-0000-0000-0000-000000000004', 'platform:bootstrap', 'Execute bootstrap operations'),
    ('10000000-0000-0000-0000-000000000005', 'tenant:manage_users', 'Manage users in a tenant'),
    ('10000000-0000-0000-0000-000000000006', 'tenant:manage_clients', 'Manage OAuth2 clients'),
    ('10000000-0000-0000-0000-000000000007', 'tenant:manage_settings', 'Manage tenant settings'),
    ('10000000-0000-0000-0000-000000000008', 'tenant:view_users', 'View users in a tenant'),
    ('10000000-0000-0000-0000-000000000009', 'tenant:view', 'View tenant metadata'),
    ('10000000-0000-0000-0000-000000000010', 'tenant:view_audit', 'View tenant audit logs'),
    ('10000000-0000-0000-0000-000000000011', 'user:read_profile', 'Read own profile'),
    ('10000000-0000-0000-0000-000000000012', 'user:write_profile', 'Update own profile'),
    ('10000000-0000-0000-0000-000000000013', 'user:change_password', 'Change own password'),
    ('10000000-0000-0000-0000-000000000014', 'user:manage_sessions', 'Manage own sessions'),
    ('10000000-0000-0000-0000-000000000015', 'client:token_introspect', 'Introspect tokens'),
    ('10000000-0000-0000-0000-000000000016', 'client:token_revoke', 'Revoke tokens')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description;

-- Seed Scoped Roles
INSERT INTO rbac_roles (id, name, scope, description) VALUES
    ('20000000-0000-0000-0000-000000000001', 'platform_admin', 'platform', 'Platform-wide administrator'),
    ('20000000-0000-0000-0000-000000000002', 'tenant_admin', 'tenant', 'Administrator for a specific tenant'),
    ('20000000-0000-0000-0000-000000000003', 'member', 'tenant', 'Regular member of a tenant')
ON CONFLICT (id) DO NOTHING;

-- Map Permissions to Roles
-- Platform Admin: All
INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT '20000000-0000-0000-0000-000000000001', id FROM rbac_permissions ON CONFLICT DO NOTHING;

-- Tenant Admin: Tenant-level management
INSERT INTO rbac_role_permissions (role_id, permission_id) VALUES
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000005'),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000006'),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000007'),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000008'),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000009'),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000010')
ON CONFLICT DO NOTHING;

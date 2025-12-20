-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Insert default tenant
INSERT INTO tenants (id, name) VALUES ('default', 'Default Tenant') ON CONFLICT (id) DO NOTHING;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES tenants(id) ON DELETE RESTRICT,
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
    
    -- Roles (from 002)
    roles JSONB DEFAULT '[]'::jsonb,

    -- Lockout (from 004)
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Ensure email is unique per tenant
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_roles ON users USING gin(roles);

-- Create credentials table (separate from users for security)
CREATE TABLE IF NOT EXISTS credentials (
    user_id VARCHAR(255) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_tenant_id ON sessions(tenant_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Create tenant_user_roles table
CREATE TABLE IF NOT EXISTS tenant_user_roles (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by VARCHAR(255) REFERENCES users(id),
    
    UNIQUE(tenant_id, user_id, role)
);

CREATE INDEX idx_tenant_user_roles_tenant_id ON tenant_user_roles(tenant_id);
CREATE INDEX idx_tenant_user_roles_user_id ON tenant_user_roles(user_id);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_credentials_updated_at BEFORE UPDATE ON credentials
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- -----------------------------------------------------------------------------
-- OAuth2 Support (Merged from 002)
-- -----------------------------------------------------------------------------

-- Create projects table
CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_projects_owner_id ON projects(owner_id);
CREATE INDEX idx_projects_deleted_at ON projects(deleted_at);

-- Create roles table for RBAC
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_permissions ON roles USING gin(permissions);

-- Create user_project_roles junction table
CREATE TABLE IF NOT EXISTS user_project_roles (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    role_id VARCHAR(255) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by VARCHAR(255) REFERENCES users(id),
    UNIQUE(user_id, project_id, role_id)
);

CREATE INDEX idx_user_project_roles_user_id ON user_project_roles(user_id);
CREATE INDEX idx_user_project_roles_project_id ON user_project_roles(project_id);
CREATE INDEX idx_user_project_roles_role_id ON user_project_roles(role_id);

-- Create OAuth2 clients table
CREATE TABLE IF NOT EXISTS oauth2_clients (
    id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    tenant_id VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES tenants(id) ON DELETE CASCADE,
    client_secret_hash TEXT NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    client_uri TEXT,
    logo_uri TEXT,
    
    -- Redirect URIs as JSONB array
    redirect_uris JSONB NOT NULL DEFAULT '[]'::jsonb,
    
    -- Allowed scopes as JSONB array
    allowed_scopes JSONB NOT NULL DEFAULT '["openid"]'::jsonb,
    
    -- Grant types as JSONB array
    grant_types JSONB NOT NULL DEFAULT '["authorization_code"]'::jsonb,
    
    -- Response types as JSONB array
    response_types JSONB NOT NULL DEFAULT '["code"]'::jsonb,
    
    -- Token settings
    token_endpoint_auth_method VARCHAR(50) DEFAULT 'client_secret_basic',
    access_token_lifetime INTEGER DEFAULT 3600,
    refresh_token_lifetime INTEGER DEFAULT 2592000,
    id_token_lifetime INTEGER DEFAULT 3600,
    
    -- Metadata
    owner_id VARCHAR(255) REFERENCES users(id) ON DELETE SET NULL,
    is_trusted BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_oauth2_clients_client_id ON oauth2_clients(client_id);
CREATE INDEX idx_oauth2_clients_owner_id ON oauth2_clients(owner_id);
CREATE INDEX idx_oauth2_clients_deleted_at ON oauth2_clients(deleted_at);
CREATE INDEX idx_oauth2_clients_redirect_uris ON oauth2_clients USING gin(redirect_uris);
CREATE INDEX idx_oauth2_clients_allowed_scopes ON oauth2_clients USING gin(allowed_scopes);

-- Create authorization codes table (short-lived)
CREATE TABLE IF NOT EXISTS authorization_codes (
    id VARCHAR(255) PRIMARY KEY,
    code VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Requested parameters
    redirect_uri TEXT NOT NULL,
    scope TEXT NOT NULL,
    state TEXT,
    nonce TEXT,
    
    -- PKCE support
    code_challenge TEXT,
    code_challenge_method VARCHAR(10),
    
    -- Expiration
    expires_at TIMESTAMP NOT NULL,
    
    -- Usage tracking
    used_at TIMESTAMP,
    is_used BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_authorization_codes_code ON authorization_codes(code);
CREATE INDEX idx_authorization_codes_client_id ON authorization_codes(client_id);
CREATE INDEX idx_authorization_codes_user_id ON authorization_codes(user_id);
CREATE INDEX idx_authorization_codes_expires_at ON authorization_codes(expires_at);
CREATE INDEX idx_authorization_codes_is_used ON authorization_codes(is_used);

-- Create access tokens table
CREATE TABLE IF NOT EXISTS access_tokens (
    id VARCHAR(255) PRIMARY KEY,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Token details
    scope TEXT NOT NULL,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    
    -- Expiration
    expires_at TIMESTAMP NOT NULL,
    
    -- Revocation
    revoked_at TIMESTAMP,
    is_revoked BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_access_tokens_token_hash ON access_tokens(token_hash);
CREATE INDEX idx_access_tokens_client_id ON access_tokens(client_id);
CREATE INDEX idx_access_tokens_user_id ON access_tokens(user_id);
CREATE INDEX idx_access_tokens_expires_at ON access_tokens(expires_at);
CREATE INDEX idx_access_tokens_is_revoked ON access_tokens(is_revoked);

-- Create refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(255) PRIMARY KEY,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    access_token_id VARCHAR(255) REFERENCES access_tokens(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Token details
    scope TEXT NOT NULL,
    
    -- Expiration
    expires_at TIMESTAMP NOT NULL,
    
    -- Revocation
    revoked_at TIMESTAMP,
    is_revoked BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_client_id ON refresh_tokens(client_id);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked);

-- Create triggers for updated_at (Merged)
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_oauth2_clients_updated_at BEFORE UPDATE ON oauth2_clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default roles
INSERT INTO roles (id, name, description, permissions) VALUES
    ('role_admin', 'admin', 'Administrator with full access', '["*"]'::jsonb),
    ('role_developer', 'developer', 'Developer with read/write access', '["read", "write", "deploy"]'::jsonb),
    ('role_viewer', 'viewer', 'Viewer with read-only access', '["read"]'::jsonb)
ON CONFLICT (name) DO NOTHING;

-- -----------------------------------------------------------------------------
-- OpenID Connect Keys (Merged from 003)
-- -----------------------------------------------------------------------------

-- Create openid_keys table
CREATE TABLE IF NOT EXISTS openid_keys (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    public_key TEXT NOT NULL,
    private_key_encrypted BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_openid_keys_expires_at ON openid_keys(expires_at);

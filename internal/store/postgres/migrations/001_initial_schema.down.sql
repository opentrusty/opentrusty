-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS openid_keys;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS access_tokens;
DROP TABLE IF EXISTS authorization_codes;
DROP TABLE IF EXISTS oauth2_clients;
DROP TABLE IF EXISTS user_project_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS tenant_user_roles;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;

-- Drop shared functions
DROP FUNCTION IF EXISTS update_updated_at_column();

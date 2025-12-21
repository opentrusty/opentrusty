-- 001_initial_schema.down.sql

DROP TABLE IF EXISTS openid_keys;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS access_tokens;
DROP TABLE IF EXISTS authorization_codes;
DROP TABLE IF EXISTS oauth2_clients;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS rbac_assignments;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS rbac_role_permissions;
DROP TABLE IF EXISTS rbac_roles;
DROP TABLE IF EXISTS rbac_permissions;

DROP FUNCTION IF EXISTS update_updated_at_column();

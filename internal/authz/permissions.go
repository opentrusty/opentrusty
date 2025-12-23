// Copyright 2026 The OpenTrusty Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

// -----------------------------------------------------------------------------
// Platform Permissions
// -----------------------------------------------------------------------------

const (
	// PermPlatformManageTenants allows creating, updating, and deleting tenants.
	PermPlatformManageTenants = "platform:manage_tenants"

	// PermPlatformManageAdmins allows assigning/revoking the platform admin role.
	PermPlatformManageAdmins = "platform:manage_admins"

	// PermPlatformViewAudit allows viewing platform-wide audit logs.
	PermPlatformViewAudit = "platform:view_audit"

	// PermPlatformBootstrap allows executing bootstrap operations.
	PermPlatformBootstrap = "platform:bootstrap"
)

// -----------------------------------------------------------------------------
// Tenant Permissions
// -----------------------------------------------------------------------------

const (
	// PermTenantManageUsers allows adding/removing users and assigning tenant roles.
	PermTenantManageUsers = "tenant:manage_users"

	// PermTenantManageClients allows registering, updating, and deleting OAuth2 clients.
	PermTenantManageClients = "tenant:manage_clients"

	// PermTenantManageSettings allows updating tenant configuration.
	PermTenantManageSettings = "tenant:manage_settings"

	// PermTenantViewUsers allows listing users and their roles within a tenant.
	PermTenantViewUsers = "tenant:view_users"

	// PermTenantView allows viewing tenant metadata.
	PermTenantView = "tenant:view"

	// PermTenantViewAudit allows viewing tenant-scoped audit logs.
	PermTenantViewAudit = "tenant:view_audit"
)

// -----------------------------------------------------------------------------
// User Permissions (Self-Service)
// -----------------------------------------------------------------------------

const (
	// PermUserReadProfile allows reading own profile information.
	PermUserReadProfile = "user:read_profile"

	// PermUserWriteProfile allows updating own profile information.
	PermUserWriteProfile = "user:write_profile"

	// PermUserChangePassword allows changing own password.
	PermUserChangePassword = "user:change_password"

	// PermUserManageSessions allows viewing and revoking own sessions.
	PermUserManageSessions = "user:manage_sessions"
)

// -----------------------------------------------------------------------------
// Client Permissions (OAuth2)
// -----------------------------------------------------------------------------

const (
	// PermClientTokenIntrospect allows introspecting access tokens.
	PermClientTokenIntrospect = "client:token_introspect"

	// PermClientTokenRevoke allows revoking tokens.
	PermClientTokenRevoke = "client:token_revoke"
)

// -----------------------------------------------------------------------------
// AllPermissions is the complete list of all defined permissions.
// Used for validation and seeding.
// -----------------------------------------------------------------------------

var AllPermissions = []string{
	// Platform
	PermPlatformManageTenants,
	PermPlatformManageAdmins,
	PermPlatformViewAudit,
	PermPlatformBootstrap,
	// Tenant
	PermTenantManageUsers,
	PermTenantManageClients,
	PermTenantManageSettings,
	PermTenantViewUsers,
	PermTenantView,
	PermTenantViewAudit,
	// User
	PermUserReadProfile,
	PermUserWriteProfile,
	PermUserChangePassword,
	PermUserManageSessions,
	// Client
	PermClientTokenIntrospect,
	PermClientTokenRevoke,
}

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
// Role Name Constants
// These are the canonical names for roles stored in the database.
// -----------------------------------------------------------------------------

const (
	// RolePlatformAdmin is the platform-wide administrator role.
	// Scope: Platform
	// Permissions: * (wildcard - all permissions)
	RolePlatformAdmin = "platform_admin"

	// RoleTenantOwner is the tenant owner role with full tenant control.
	// Scope: Tenant
	RoleTenantOwner = "tenant_owner"

	// RoleTenantAdmin is the tenant administrator role.
	// Scope: Tenant
	RoleTenantAdmin = "tenant_admin"

	// RoleTenantMember is a basic tenant membership role.
	// Scope: Tenant
	RoleTenantMember = "tenant_member"
)

// -----------------------------------------------------------------------------
// Actor Type Constants
// These identify the type of actor making a request.
// -----------------------------------------------------------------------------

type ActorType string

const (
	// ActorUser represents a human user.
	ActorUser ActorType = "user"

	// ActorClient represents an OAuth2 client acting on behalf of a user.
	ActorClient ActorType = "client"

	// ActorSystem represents internal system operations (e.g., bootstrap, scheduled jobs).
	ActorSystem ActorType = "system"
)

// -----------------------------------------------------------------------------
// Role Permission Mappings
// These define the default permissions for each role.
// Used for seeding and validation.
// -----------------------------------------------------------------------------

// PlatformAdminPermissions defines permissions for the platform_admin role.
var PlatformAdminPermissions = []string{
	"*", // Wildcard: all permissions
}

// TenantOwnerPermissions defines permissions for the tenant_owner role.
var TenantOwnerPermissions = []string{
	PermTenantManageUsers,
	PermTenantManageClients,
	PermTenantManageSettings,
	PermTenantViewUsers,
	PermTenantView,
	PermTenantViewAudit,
	PermUserReadProfile,
	PermUserWriteProfile,
	PermUserChangePassword,
	PermUserManageSessions,
}

// TenantAdminPermissions defines permissions for the tenant_admin role.
var TenantAdminPermissions = []string{
	PermTenantManageUsers,
	PermTenantManageClients,
	PermTenantViewUsers,
	PermTenantView,
	PermUserReadProfile,
	PermUserWriteProfile,
	PermUserChangePassword,
	PermUserManageSessions,
}

// TenantMemberPermissions defines permissions for the tenant_member role.
var TenantMemberPermissions = []string{
	PermTenantView,
	PermUserReadProfile,
	PermUserWriteProfile,
	PermUserChangePassword,
}

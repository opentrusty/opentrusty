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

package rbac

// System-defined Role IDs from initial schema migration (001_initial_schema.up.sql).
// These UUIDs are seeded during database initialization and must remain stable.
// DO NOT modify these values without corresponding migration and data migration plan.
const (
	// RoleIDPlatformAdmin grants platform-wide administrative privileges.
	// Scope: platform (scope_context_id = NULL)
	// Permissions: All permissions (via wildcard mapping in migration)
	RoleIDPlatformAdmin = "20000000-0000-0000-0000-000000000001"

	// RoleIDTenantAdmin grants administrative privileges within a specific tenant.
	// Scope: tenant (scope_context_id = tenant UUID)
	// Permissions: user:provision, user:manage, client:register
	RoleIDTenantAdmin = "20000000-0000-0000-0000-000000000002"

	// RoleIDMember grants basic member privileges within a specific tenant.
	// Scope: tenant (scope_context_id = tenant UUID)
	// Permissions: Currently none (to be extended)
	RoleIDMember = "20000000-0000-0000-0000-000000000003"
)

// System-defined Permission IDs from initial schema migration (001_initial_schema.up.sql).
// These UUIDs are seeded during database initialization and must remain stable.
const (
	// PermissionIDTenantCreate allows creating new tenants (platform-level).
	PermissionIDTenantCreate = "10000000-0000-0000-0000-000000000001"

	// PermissionIDTenantDelete allows deleting tenants (platform-level).
	PermissionIDTenantDelete = "10000000-0000-0000-0000-000000000002"

	// PermissionIDTenantList allows listing all tenants (platform-level).
	PermissionIDTenantList = "10000000-0000-0000-0000-000000000003"

	// PermissionIDUserProvision allows provisioning users within a tenant.
	PermissionIDUserProvision = "10000000-0000-0000-0000-000000000004"

	// PermissionIDUserManage allows managing users within a tenant.
	PermissionIDUserManage = "10000000-0000-0000-0000-000000000005"

	// PermissionIDClientRegister allows registering OAuth2 clients within a tenant.
	PermissionIDClientRegister = "10000000-0000-0000-0000-000000000006"
)

// SystemTenantID is the pre-seeded tenant used for initial platform admin bootstrap.
// This tenant is created during database migration and should not be deleted.
const SystemTenantID = "10000000-0000-0000-0000-000000000000"

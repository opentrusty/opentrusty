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

package authz_test

import (
	"context"
	"testing"

	"github.com/opentrusty/opentrusty/internal/authz"
)

// MockRoleRepository implements authz.RoleRepository for testing
type MockRoleRepository struct {
	roles map[string]*authz.Role
}

func NewMockRoleRepository() *MockRoleRepository {
	return &MockRoleRepository{
		roles: map[string]*authz.Role{
			"platform_admin": {
				ID:          "role-platform-admin",
				Name:        authz.RolePlatformAdmin,
				Scope:       authz.ScopePlatform,
				Permissions: authz.PlatformAdminPermissions,
			},
			"tenant_owner": {
				ID:          "role-tenant-owner",
				Name:        authz.RoleTenantOwner,
				Scope:       authz.ScopeTenant,
				Permissions: authz.TenantOwnerPermissions,
			},
			"tenant_admin": {
				ID:          "role-tenant-admin",
				Name:        authz.RoleTenantAdmin,
				Scope:       authz.ScopeTenant,
				Permissions: authz.TenantAdminPermissions,
			},
			"tenant_member": {
				ID:          "role-tenant-member",
				Name:        authz.RoleTenantMember,
				Scope:       authz.ScopeTenant,
				Permissions: authz.TenantMemberPermissions,
			},
		},
	}
}

func (m *MockRoleRepository) Create(role *authz.Role) error { return nil }
func (m *MockRoleRepository) GetByID(id string) (*authz.Role, error) {
	for _, r := range m.roles {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, authz.ErrRoleNotFound
}
func (m *MockRoleRepository) GetByName(name string, scope authz.Scope) (*authz.Role, error) {
	if r, ok := m.roles[name]; ok && r.Scope == scope {
		return r, nil
	}
	return nil, authz.ErrRoleNotFound
}
func (m *MockRoleRepository) Update(role *authz.Role) error                  { return nil }
func (m *MockRoleRepository) Delete(id string) error                         { return nil }
func (m *MockRoleRepository) List(scope *authz.Scope) ([]*authz.Role, error) { return nil, nil }

// MockAssignmentRepository implements authz.AssignmentRepository for testing
type MockAssignmentRepository struct {
	assignments []*authz.Assignment
}

func NewMockAssignmentRepository() *MockAssignmentRepository {
	return &MockAssignmentRepository{
		assignments: []*authz.Assignment{},
	}
}

func (m *MockAssignmentRepository) Grant(a *authz.Assignment) error {
	m.assignments = append(m.assignments, a)
	return nil
}
func (m *MockAssignmentRepository) Revoke(userID, roleID string, scope authz.Scope, scopeContextID *string) error {
	return nil
}
func (m *MockAssignmentRepository) ListForUser(userID string) ([]*authz.Assignment, error) {
	var result []*authz.Assignment
	for _, a := range m.assignments {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result, nil
}
func (m *MockAssignmentRepository) ListByRole(roleID string, scope authz.Scope, scopeContextID *string) ([]string, error) {
	return nil, nil
}
func (m *MockAssignmentRepository) CheckExists(roleID string, scope authz.Scope, scopeContextID *string) (bool, error) {
	return false, nil
}

// MockProjectRepository implements authz.ProjectRepository for testing
type MockProjectRepository struct{}

func (m *MockProjectRepository) Create(project *authz.Project) error           { return nil }
func (m *MockProjectRepository) GetByID(id string) (*authz.Project, error)     { return nil, nil }
func (m *MockProjectRepository) GetByName(name string) (*authz.Project, error) { return nil, nil }
func (m *MockProjectRepository) Update(project *authz.Project) error           { return nil }
func (m *MockProjectRepository) Delete(id string) error                        { return nil }
func (m *MockProjectRepository) ListByOwner(ownerID string) ([]*authz.Project, error) {
	return nil, nil
}
func (m *MockProjectRepository) ListByUser(userID string) ([]*authz.Project, error) { return nil, nil }

// TestPurpose: Validates that Platform Admin privileges are scoped to the platform and strictly completely isolated from Tenant Admin privileges where appropriate (and vice versa).
// Scope: Unit Test
// Security: RBAC Scope Isolation (prevents vertical privilege escalation)
// Permissions: platform:manage_tenants
// Expected: Platform Admin has platform permissions, Tenant Admin does not.
// Test Case ID: AUT-03
func TestAuthz_Scope_PlatformAdminVsTenantAdmin(t *testing.T) {
	roleRepo := NewMockRoleRepository()
	assignmentRepo := NewMockAssignmentRepository()
	projectRepo := &MockProjectRepository{}

	svc := authz.NewService(projectRepo, roleRepo, assignmentRepo)
	ctx := context.Background()

	tenantID := "tenant-123"

	// Setup: User A is platform admin
	assignmentRepo.Grant(&authz.Assignment{
		ID:             "assign-1",
		UserID:         "user-platform-admin",
		RoleID:         "role-platform-admin",
		Scope:          authz.ScopePlatform,
		ScopeContextID: nil,
	})

	// Setup: User B is tenant admin
	assignmentRepo.Grant(&authz.Assignment{
		ID:             "assign-2",
		UserID:         "user-tenant-admin",
		RoleID:         "role-tenant-admin",
		Scope:          authz.ScopeTenant,
		ScopeContextID: &tenantID,
	})

	// Test: Platform admin can manage tenants
	allowed, err := svc.HasPermission(ctx, "user-platform-admin", authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("platform admin should have PermPlatformManageTenants")
	}

	// Test: Tenant admin cannot manage tenants (platform permission)
	allowed, err = svc.HasPermission(ctx, "user-tenant-admin", authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("tenant admin should NOT have PermPlatformManageTenants")
	}
}

// TestPurpose: Validates that Tenant Admin privileges are correctly authorized while Tenant Member privileges are restricted within the same tenant.
// Scope: Unit Test
// Security: RBAC Permission Enforcement
// Permissions: tenant:manage_users, tenant:view
// Expected: Tenant Admin has management permissions, Tenant Member only has view permissions.
// Test Case ID: AUT-04
func TestAuthz_Scope_TenantAdminVsTenantMember(t *testing.T) {
	roleRepo := NewMockRoleRepository()
	assignmentRepo := NewMockAssignmentRepository()
	projectRepo := &MockProjectRepository{}

	svc := authz.NewService(projectRepo, roleRepo, assignmentRepo)
	ctx := context.Background()

	tenantID := "tenant-123"

	// Setup: User A is tenant admin
	assignmentRepo.Grant(&authz.Assignment{
		ID:             "assign-1",
		UserID:         "user-tenant-admin",
		RoleID:         "role-tenant-admin",
		Scope:          authz.ScopeTenant,
		ScopeContextID: &tenantID,
	})

	// Setup: User B is tenant member
	assignmentRepo.Grant(&authz.Assignment{
		ID:             "assign-2",
		UserID:         "user-tenant-member",
		RoleID:         "role-tenant-member",
		Scope:          authz.ScopeTenant,
		ScopeContextID: &tenantID,
	})

	// Test: Tenant admin can manage users
	allowed, err := svc.HasPermission(ctx, "user-tenant-admin", authz.ScopeTenant, &tenantID, authz.PermTenantManageUsers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("tenant admin should have PermTenantManageUsers")
	}

	// Test: Tenant member cannot manage users
	allowed, err = svc.HasPermission(ctx, "user-tenant-member", authz.ScopeTenant, &tenantID, authz.PermTenantManageUsers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("tenant member should NOT have PermTenantManageUsers")
	}

	// Test: Both can view tenant
	allowed, _ = svc.HasPermission(ctx, "user-tenant-admin", authz.ScopeTenant, &tenantID, authz.PermTenantView)
	if !allowed {
		t.Error("tenant admin should have PermTenantView")
	}

	allowed, _ = svc.HasPermission(ctx, "user-tenant-member", authz.ScopeTenant, &tenantID, authz.PermTenantView)
	if !allowed {
		t.Error("tenant member should have PermTenantView")
	}
}

// TestPurpose: Validates that a user with admin privileges in Tenant A cannot perform actions in Tenant B.
// Scope: Unit Test
// Security: Multi-tenancy Data Isolation (prevents lateral movement / horizontal privilege escalation)
// Expected: Access to Tenant B resources is denied for Tenant A admin.
// Test Case ID: AUT-05
func TestAuthz_Isolation_CrossTenantAccessDenied(t *testing.T) {
	roleRepo := NewMockRoleRepository()
	assignmentRepo := NewMockAssignmentRepository()
	projectRepo := &MockProjectRepository{}

	svc := authz.NewService(projectRepo, roleRepo, assignmentRepo)
	ctx := context.Background()

	tenantA := "tenant-A"
	tenantB := "tenant-B"

	// Setup: User is admin of Tenant A only
	assignmentRepo.Grant(&authz.Assignment{
		ID:             "assign-1",
		UserID:         "user-admin-A",
		RoleID:         "role-tenant-admin",
		Scope:          authz.ScopeTenant,
		ScopeContextID: &tenantA,
	})

	// Test: User can manage users in Tenant A
	allowed, err := svc.HasPermission(ctx, "user-admin-A", authz.ScopeTenant, &tenantA, authz.PermTenantManageUsers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("user should have permission in Tenant A")
	}

	// Test: User CANNOT manage users in Tenant B (cross-tenant denial)
	allowed, err = svc.HasPermission(ctx, "user-admin-A", authz.ScopeTenant, &tenantB, authz.PermTenantManageUsers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("user should NOT have permission in Tenant B - cross-tenant access must be denied")
	}
}

// TestPurpose: Validates the HasPermission method on the Role struct for various scenarios (wildcards, direct matches).
// Scope: Unit Test
// Security: Core logic for permission checking
// Expected: Returns true if role has permission, false otherwise.
// Test Case ID: AUT-06
func TestAuthz_Role_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       *authz.Role
		permission string
		expected   bool
	}{
		{
			name:       "platform admin has wildcard",
			role:       &authz.Role{Permissions: []string{"*"}},
			permission: authz.PermPlatformManageTenants,
			expected:   true,
		},
		{
			name:       "tenant admin has manage users",
			role:       &authz.Role{Permissions: authz.TenantAdminPermissions},
			permission: authz.PermTenantManageUsers,
			expected:   true,
		},
		{
			name:       "tenant member does not have manage users",
			role:       &authz.Role{Permissions: authz.TenantMemberPermissions},
			permission: authz.PermTenantManageUsers,
			expected:   false,
		},
		{
			name:       "tenant member has view",
			role:       &authz.Role{Permissions: authz.TenantMemberPermissions},
			permission: authz.PermTenantView,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.HasPermission(tt.permission)
			if result != tt.expected {
				t.Errorf("HasPermission(%q) = %v, want %v", tt.permission, result, tt.expected)
			}
		})
	}
}

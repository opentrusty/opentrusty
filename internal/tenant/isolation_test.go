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

package tenant

import (
	"context"
	"errors"
	"testing"

	"github.com/opentrusty/opentrusty/internal/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRoleRepo implements RoleRepository for testing
type mockRoleRepo struct {
	mock.Mock
}

func (m *mockRoleRepo) AssignRole(ctx context.Context, role *TenantUserRole) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *mockRoleRepo) RevokeRole(ctx context.Context, tenantID, userID, role string) error {
	args := m.Called(ctx, tenantID, userID, role)
	return args.Error(0)
}

func (m *mockRoleRepo) GetUserRoles(ctx context.Context, tenantID, userID string) ([]*TenantUserRole, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TenantUserRole), args.Error(1)
}

func (m *mockRoleRepo) GetTenantUsers(ctx context.Context, tenantID string) ([]*TenantUserRole, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TenantUserRole), args.Error(1)
}

// ErrInvalidRole for test validation
var ErrInvalidRoleTest = errors.New("invalid role")

// TestPurpose: Validates that tenant-scoped operations strictly require a non-empty tenant ID to prevent global data exposure.
// Scope: Unit Test
// Security: Multi-tenant boundary enforcement
// Expected: Returns an error when an empty tenant ID is provided.
// Test Case ID: TEN-02
// RelatedSpecs: Multi-tenant Logical Separation (consistent with NIST Cloud model)
func TestTenant_Isolation_TenantIDMustBePresent(t *testing.T) {
	repo := new(mockRepo)
	roleRepo := new(mockRoleRepo)
	authzRepo := new(mockAssignmentRepo)
	auditLogger := &mockAudit{}
	auditLogger.On("Log", mock.Anything, mock.Anything).Return()

	service := NewService(repo, roleRepo, authzRepo, auditLogger)
	ctx := context.Background()

	// Test case: Empty tenant ID should fail
	t.Run("GetTenant_EmptyTenantID_ReturnsError", func(t *testing.T) {
		repo.On("GetByID", ctx, "").Return((*Tenant)(nil), errors.New("invalid empty tenant ID"))

		_, err := service.GetTenant(ctx, "")
		assert.Error(t, err, "empty tenant ID should fail")
	})
}

// TestPurpose: Validates that defined role constants are correctly accepted for assignment.
// Scope: Unit Test
// Security: RBAC data integrity
// Expected: Assignment succeeds for valid role constants (owner, admin, member).
// Test Case ID: TEN-03
func TestTenant_Isolation_AssignValidRole_Succeeds(t *testing.T) {
	repo := new(mockRepo)
	roleRepo := new(mockRoleRepo)
	authzRepo := new(mockAssignmentRepo)
	auditLogger := &mockAudit{}
	auditLogger.On("Log", mock.Anything, mock.Anything).Return()

	service := NewService(repo, roleRepo, authzRepo, auditLogger)
	ctx := context.Background()

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	grantedBy := id.NewUUIDv7()

	roleRepo.On("AssignRole", ctx, mock.MatchedBy(func(r *TenantUserRole) bool {
		return r.TenantID == tenantID && r.UserID == userID && r.Role == RoleTenantAdmin
	})).Return(nil)

	err := service.AssignRole(ctx, tenantID, userID, RoleTenantAdmin, grantedBy)
	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

// TestPurpose: Validates that non-defined role names are rejected to prevent arbitrary privilege assignment.
// Scope: Unit Test
// Security: Unauthorized privilege escalation prevention
// Expected: Returns an error for role names not in the allowed list.
// Test Case ID: TEN-04
func TestTenant_Isolation_AssignInvalidRole_ReturnsError(t *testing.T) {
	repo := new(mockRepo)
	roleRepo := new(mockRoleRepo)
	authzRepo := new(mockAssignmentRepo)
	auditLogger := &mockAudit{}
	auditLogger.On("Log", mock.Anything, mock.Anything).Return()

	service := NewService(repo, roleRepo, authzRepo, auditLogger)
	ctx := context.Background()

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	grantedBy := id.NewUUIDv7()

	// "super_admin" is not a valid tenant role
	err := service.AssignRole(ctx, tenantID, userID, "super_admin", grantedBy)
	assert.Error(t, err, "invalid role should be rejected")
}

// TestPurpose: Validates that role revocation works for valid roles.
// Scope: Unit Test
// Security: Revocation enforcement
// Expected: Role is successfully revoked for the user in the tenant.
// Test Case ID: TEN-05
func TestTenant_Isolation_RevokeRole_ValidRole_Succeeds(t *testing.T) {
	repo := new(mockRepo)
	roleRepo := new(mockRoleRepo)
	authzRepo := new(mockAssignmentRepo)
	auditLogger := &mockAudit{}
	auditLogger.On("Log", mock.Anything, mock.Anything).Return()

	service := NewService(repo, roleRepo, authzRepo, auditLogger)
	ctx := context.Background()

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()

	roleRepo.On("RevokeRole", ctx, tenantID, userID, RoleTenantMember).Return(nil)

	err := service.RevokeRole(ctx, tenantID, userID, RoleTenantMember)
	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

// TestPurpose: Validates that retrieval of user roles returns all roles assigned to that user in a specific tenant.
// Scope: Unit Test
// Security: RBAC role transparency
// Expected: Returns a list of all assigned roles for the user.
// Test Case ID: TEN-06
func TestTenant_Isolation_GetUserRoles_ReturnsAllRoles(t *testing.T) {
	repo := new(mockRepo)
	roleRepo := new(mockRoleRepo)
	authzRepo := new(mockAssignmentRepo)
	auditLogger := &mockAudit{}

	service := NewService(repo, roleRepo, authzRepo, auditLogger)
	ctx := context.Background()

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()

	expectedRoles := []*TenantUserRole{
		{TenantID: tenantID, UserID: userID, Role: RoleTenantAdmin},
		{TenantID: tenantID, UserID: userID, Role: RoleTenantMember},
	}

	roleRepo.On("GetUserRoles", ctx, tenantID, userID).Return(expectedRoles, nil)

	roles, err := service.GetUserRoles(ctx, tenantID, userID)
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	roleRepo.AssertExpectations(t)
}

// TestPurpose: Validates that only defined tenant role constants (RoleTenantOwner, RoleTenantAdmin, RoleTenantMember) are accepted.
// Scope: Unit Test
// Security: Role name validation logic
// Expected: Accepts defined constants, rejects anything else.
// Test Case ID: TEN-07
func TestTenant_Isolation_RoleValidation_OnlyAcceptsDefinedConstants(t *testing.T) {
	repo := new(mockRepo)
	roleRepo := new(mockRoleRepo)
	authzRepo := new(mockAssignmentRepo)
	auditLogger := &mockAudit{}
	auditLogger.On("Log", mock.Anything, mock.Anything).Return()

	service := NewService(repo, roleRepo, authzRepo, auditLogger)
	ctx := context.Background()

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	grantedBy := id.NewUUIDv7()

	// Valid roles should succeed
	validRoles := []string{RoleTenantOwner, RoleTenantAdmin, RoleTenantMember}
	for _, role := range validRoles {
		roleRepo.On("AssignRole", ctx, mock.MatchedBy(func(r *TenantUserRole) bool {
			return r.Role == role
		})).Return(nil).Once()
		err := service.AssignRole(ctx, tenantID, userID, role, grantedBy)
		assert.NoError(t, err, "valid role %s should be accepted", role)
	}

	// Invalid roles should fail
	invalidRoles := []string{"admin", "root", "platform_admin", ""}
	for _, role := range invalidRoles {
		err := service.AssignRole(ctx, tenantID, userID, role, grantedBy)
		assert.Error(t, err, "invalid role %s should be rejected", role)
	}
}

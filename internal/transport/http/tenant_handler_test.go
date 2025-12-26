package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repositories for Authz
type mockAssignmentRepo struct {
	mock.Mock
}

func (m *mockAssignmentRepo) Grant(a *authz.Assignment) error                     { return nil }
func (m *mockAssignmentRepo) Revoke(u, r string, s authz.Scope, sc *string) error { return nil }
func (m *mockAssignmentRepo) ListForUser(userID string) ([]*authz.Assignment, error) {
	args := m.Called(userID)
	return args.Get(0).([]*authz.Assignment), args.Error(1)
}
func (m *mockAssignmentRepo) ListByRole(r string, s authz.Scope, sc *string) ([]string, error) {
	return nil, nil
}
func (m *mockAssignmentRepo) CheckExists(r string, s authz.Scope, sc *string) (bool, error) {
	return false, nil
}

type mockAuthzRoleRepo struct {
	mock.Mock
}

func (m *mockAuthzRoleRepo) Create(r *authz.Role) error { return nil }
func (m *mockAuthzRoleRepo) GetByID(id string) (*authz.Role, error) {
	args := m.Called(id)
	return args.Get(0).(*authz.Role), args.Error(1)
}
func (m *mockAuthzRoleRepo) GetByName(n string, s authz.Scope) (*authz.Role, error) { return nil, nil }
func (m *mockAuthzRoleRepo) Update(r *authz.Role) error                             { return nil }
func (m *mockAuthzRoleRepo) Delete(id string) error                                 { return nil }
func (m *mockAuthzRoleRepo) List(s *authz.Scope) ([]*authz.Role, error)             { return nil, nil }

// Mock Repository for Tenant
type mockTenantRepo struct {
	mock.Mock
}

func (m *mockTenantRepo) Create(ctx context.Context, t *tenant.Tenant) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}
func (m *mockTenantRepo) GetByID(ctx context.Context, id string) (*tenant.Tenant, error) {
	return nil, nil
}
func (m *mockTenantRepo) GetByName(ctx context.Context, name string) (*tenant.Tenant, error) {
	return nil, nil
}
func (m *mockTenantRepo) Update(ctx context.Context, t *tenant.Tenant) error { return nil }
func (m *mockTenantRepo) Delete(ctx context.Context, id string) error        { return nil }
func (m *mockTenantRepo) List(ctx context.Context, l, o int) ([]*tenant.Tenant, error) {
	return nil, nil
}

// TestPurpose: Validates authorization rules for creating tenants (only platform admins).
// Scope: Unit Test
// Security: RBAC enforcement (prevents unauthorized tenant creation)
// Permissions: platform:manage_tenants
// Expected: Returns HTTP 201 Created for platform admins, and 403 Forbidden for others.
// Test Case ID: TEN-07
func TestTenant_Create_AuthorizationEnforcement(t *testing.T) {
	// 1. Setup
	assignRepo := new(mockAssignmentRepo)
	authzRoleRepo := new(mockAuthzRoleRepo)
	authzSvc := authz.NewService(nil, authzRoleRepo, assignRepo)

	tenantRepo := new(mockTenantRepo)
	tenantSvc := tenant.NewService(tenantRepo, nil, assignRepo, audit.NewSlogLogger())

	h := &Handler{
		authzService:  authzSvc,
		tenantService: tenantSvc,
		auditLogger:   audit.NewSlogLogger(),
	}

	t.Run("Forbidden for non-admin", func(t *testing.T) {
		userID := "user-123"
		assignRepo.On("ListForUser", userID).Return([]*authz.Assignment{}, nil)

		reqBody, _ := json.Marshal(CreateTenantRequest{Name: "New Tenant"})
		req := httptest.NewRequest("POST", "/tenants", bytes.NewReader(reqBody))
		ctx := context.WithValue(req.Context(), userIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CreateTenant(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "platform admin")
	})

	t.Run("Success for platform admin", func(t *testing.T) {
		userID := "admin-123"
		roleID := "role-admin"

		// Setup Role with permission
		authzRoleRepo.On("GetByID", roleID).Return(&authz.Role{
			ID:          roleID,
			Name:        "Platform Admin",
			Permissions: []string{authz.PermPlatformManageTenants},
		}, nil)

		// Setup Assignment
		assignRepo.On("ListForUser", userID).Return([]*authz.Assignment{
			{
				UserID: userID,
				RoleID: roleID,
				Scope:  authz.ScopePlatform,
			},
		}, nil)

		// Setup Tenant Creation
		tenantRepo.On("Create", mock.Anything, mock.MatchedBy(func(ten *tenant.Tenant) bool {
			return ten.Name == "New Tenant" && ten.ID != ""
		})).Return(nil)

		reqBody, _ := json.Marshal(CreateTenantRequest{Name: "New Tenant"})
		req := httptest.NewRequest("POST", "/tenants", bytes.NewReader(reqBody))
		ctx := context.WithValue(req.Context(), userIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CreateTenant(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp tenant.Tenant
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "New Tenant", resp.Name)
		assert.NotEmpty(t, resp.ID)
	})
}

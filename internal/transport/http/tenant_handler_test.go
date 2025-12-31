package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/identity"
	"github.com/opentrusty/opentrusty/internal/oauth2"
	"github.com/opentrusty/opentrusty/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	os.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	os.Exit(m.Run())
}

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
func (m *mockAssignmentRepo) DeleteByContextID(s authz.Scope, cid string) error { return nil }

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

// Mock Repository for TenantRole
type mockTenantRoleRepo struct {
	mock.Mock
}

func (m *mockTenantRoleRepo) AssignRole(ctx context.Context, r *tenant.TenantUserRole) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *mockTenantRoleRepo) RevokeRole(ctx context.Context, tenantID, userID, role string) error {
	return nil
}
func (m *mockTenantRoleRepo) GetUserRoles(ctx context.Context, tenantID, userID string) ([]*tenant.TenantUserRole, error) {
	return nil, nil
}
func (m *mockTenantRoleRepo) GetTenantUsers(ctx context.Context, tenantID string) ([]*tenant.TenantUserRole, error) {
	return nil, nil
}
func (m *mockTenantRoleRepo) DeleteByTenantID(ctx context.Context, tenantID string) error { return nil }

type mockClientRepo struct {
	mock.Mock
}

func (m *mockClientRepo) Create(c *oauth2.Client) error                         { return nil }
func (m *mockClientRepo) GetByID(id string, tID string) (*oauth2.Client, error) { return nil, nil }
func (m *mockClientRepo) GetByClientID(id string, tID string) (*oauth2.Client, error) {
	return nil, nil
}
func (m *mockClientRepo) Update(c *oauth2.Client) error                     { return nil }
func (m *mockClientRepo) Delete(id string, tID string) error                { return nil }
func (m *mockClientRepo) ListByOwner(oID string) ([]*oauth2.Client, error)  { return nil, nil }
func (m *mockClientRepo) ListByTenant(tID string) ([]*oauth2.Client, error) { return nil, nil }
func (m *mockClientRepo) DeleteByTenantID(tID string) error                 { return nil }

type mockMembershipRepo struct {
	mock.Mock
}

func (m *mockMembershipRepo) Create(ctx context.Context, mem *tenant.Membership) error {
	args := m.Called(ctx, mem)
	return args.Error(0)
}
func (m *mockMembershipRepo) Delete(ctx context.Context, tenantID, userID string) error {
	return nil
}
func (m *mockMembershipRepo) ListByUser(ctx context.Context, userID string) ([]*tenant.Membership, error) {
	return nil, nil
}
func (m *mockMembershipRepo) ListByTenant(ctx context.Context, tenantID string) ([]*tenant.Membership, error) {
	return nil, nil
}
func (m *mockMembershipRepo) DeleteByTenantID(ctx context.Context, tenantID string) error {
	return nil
}

// Mock Repository for Identity
type mockIdentityRepo struct {
	mock.Mock
}

func (m *mockIdentityRepo) Create(u *identity.User) error {
	args := m.Called(u)
	return args.Error(0)
}
func (m *mockIdentityRepo) GetByID(id string) (*identity.User, error) {
	args := m.Called(id)
	return args.Get(0).(*identity.User), args.Error(1)
}
func (m *mockIdentityRepo) GetByEmail(email string) (*identity.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, identity.ErrUserNotFound
	}
	return args.Get(0).(*identity.User), args.Error(1)
}
func (m *mockIdentityRepo) Update(u *identity.User) error                           { return nil }
func (m *mockIdentityRepo) UpdateLockout(id string, a int, until *time.Time) error  { return nil }
func (m *mockIdentityRepo) Delete(id string) error                                  { return nil }
func (m *mockIdentityRepo) DeleteByTenantID(tenantID string) error                  { return nil }
func (m *mockIdentityRepo) GetCredentials(id string) (*identity.Credentials, error) { return nil, nil }
func (m *mockIdentityRepo) AddCredentials(c *identity.Credentials) error {
	args := m.Called(c)
	return args.Error(0)
}
func (m *mockIdentityRepo) UpdatePassword(id string, h string) error { return nil }

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
	tenantRoleRepo := new(mockTenantRoleRepo)
	identityRepo := new(mockIdentityRepo)
	clientRepo := new(mockClientRepo)
	membershipRepo := new(mockMembershipRepo)

	hasher := identity.NewPasswordHasher(64*1024, 1, 1, 16, 32)
	identitySvc := identity.NewService(identityRepo, hasher, audit.NewSlogLogger(), 5, 15*time.Minute)
	tenantSvc := tenant.NewService(tenantRepo, tenantRoleRepo, assignRepo, identitySvc, clientRepo, membershipRepo, audit.NewSlogLogger())

	h := NewHandler(
		identitySvc,
		nil,
		nil,
		authzSvc,
		tenantSvc,
		nil,
		audit.NewSlogLogger(),
		nil,
		SessionConfig{},
		[]byte("test-signing-key"),
		"admin",
	)

	t.Run("Forbidden for non-admin", func(t *testing.T) {
		userID := "user-123"
		assignRepo.On("ListForUser", userID).Return([]*authz.Assignment{}, nil)

		reqBody, _ := json.Marshal(CreateTenantRequest{Name: "New Tenant", AdminEmail: "admin@example.com"})
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

		// Expectations for owner provisioning (now inside tenantService.CreateTenant)
		identityRepo.On("GetByEmail", "admin@example.com").Return((*identity.User)(nil), identity.ErrUserNotFound)
		identityRepo.On("Create", mock.Anything).Return(nil)
		identityRepo.On("AddCredentials", mock.Anything).Return(nil)
		membershipRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		tenantRoleRepo.On("AssignRole", mock.Anything, mock.Anything).Return(nil)

		reqBody, _ := json.Marshal(CreateTenantRequest{Name: "New Tenant", AdminEmail: "admin@example.com"})
		req := httptest.NewRequest("POST", "/tenants", bytes.NewReader(reqBody))
		ctx := context.WithValue(req.Context(), userIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CreateTenant(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "New Tenant", resp["name"])
		assert.NotEmpty(t, resp["id"])
	})
}

func toPtr(s string) *string {
	return &s
}

package tenant

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

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

func (m *mockIdentityRepo) Update(u *identity.User) error                          { return nil }
func (m *mockIdentityRepo) UpdateLockout(id string, a int, until *time.Time) error { return nil }
func (m *mockIdentityRepo) Delete(id string) error                                 { return nil }
func (m *mockIdentityRepo) DeleteByTenantID(tenantID string) error                 { return nil }

func (m *mockIdentityRepo) GetCredentials(id string) (*identity.Credentials, error) {
	return nil, nil
}

func (m *mockIdentityRepo) AddCredentials(c *identity.Credentials) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *mockIdentityRepo) UpdatePassword(id string, h string) error { return nil }

func (m *mockRepo) Create(ctx context.Context, t *Tenant) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Tenant, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Tenant), args.Error(1)
}

func (m *mockRepo) GetByName(ctx context.Context, name string) (*Tenant, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*Tenant), args.Error(1)
}

func (m *mockRepo) Update(ctx context.Context, t *Tenant) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]*Tenant, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*Tenant), args.Error(1)
}

type mockAssignmentRepo struct {
	mock.Mock
}

func (m *mockAssignmentRepo) Grant(assignment *authz.Assignment) error {
	args := m.Called(assignment)
	return args.Error(0)
}

func (m *mockAssignmentRepo) Revoke(userID, roleID string, scope authz.Scope, scopeContextID *string) error {
	args := m.Called(userID, roleID, scope, scopeContextID)
	return args.Error(0)
}

func (m *mockAssignmentRepo) ListForUser(userID string) ([]*authz.Assignment, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authz.Assignment), args.Error(1)
}

func (m *mockAssignmentRepo) ListByRole(roleID string, scope authz.Scope, scopeContextID *string) ([]string, error) {
	args := m.Called(roleID, scope, scopeContextID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockAssignmentRepo) CheckExists(roleID string, scope authz.Scope, scopeContextID *string) (bool, error) {
	args := m.Called(roleID, scope, scopeContextID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *mockAssignmentRepo) DeleteByContextID(scope authz.Scope, contextID string) error {
	args := m.Called(scope, contextID)
	return args.Error(0)
}

type mockAudit struct {
	mock.Mock
}

func (m *mockAudit) Log(ctx context.Context, event audit.Event) {
	m.Called(ctx, event)
}

type mockMembershipRepo struct {
	mock.Mock
}

func (m *mockMembershipRepo) Create(ctx context.Context, mem *Membership) error {
	args := m.Called(ctx, mem)
	return args.Error(0)
}

func (m *mockMembershipRepo) Delete(ctx context.Context, tenantID, userID string) error {
	args := m.Called(ctx, tenantID, userID)
	return args.Error(0)
}

func (m *mockMembershipRepo) ListByUser(ctx context.Context, userID string) ([]*Membership, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*Membership), args.Error(1)
}

func (m *mockMembershipRepo) ListByTenant(ctx context.Context, tenantID string) ([]*Membership, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*Membership), args.Error(1)
}

func (m *mockMembershipRepo) DeleteByTenantID(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

// TestPurpose: Validates that tenant creation correctly generates IDs using UUIDv7 for temporal ordering.
// Scope: Unit Test
// Security: Traceability and unique identification of tenants
// Expected: A new tenant is created with a valid UUIDv7 ID and the provided name.
// Test Case ID: TEN-01
func TestTenant_Service_CreateTenant_UUIDv7(t *testing.T) {
	repo := new(mockRepo)
	roleRepo := new(mockRoleRepo)
	authzRepo := new(mockAssignmentRepo)
	membershipRepo := new(mockMembershipRepo)
	auditLogger := new(mockAudit)
	name := "Test Tenant"
	creatorID := "user-123"
	ctx := context.Background()

	auditLogger.On("Log", mock.Anything, mock.Anything).Return()

	repo.On("GetByName", ctx, name).Return((*Tenant)(nil), nil)

	repo.On("Create", ctx, mock.MatchedBy(func(t *Tenant) bool {
		return t.Name == name
	})).Return(nil)

	// Mock owner provisioning
	identityRepo := new(mockIdentityRepo)
	hasher := identity.NewPasswordHasher(64*1024, 1, 1, 16, 32)
	identitySvc := identity.NewService(identityRepo, hasher, auditLogger, 5, 15*time.Minute)
	service := NewService(repo, roleRepo, authzRepo, identitySvc, nil, membershipRepo, auditLogger)

	identityRepo.On("GetByEmail", "admin@example.com").Return((*identity.User)(nil), identity.ErrUserNotFound)
	identityRepo.On("Create", mock.Anything).Return(nil)
	identityRepo.On("AddCredentials", mock.Anything).Return(nil)
	membershipRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	roleRepo.On("AssignRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	authzRepo.On("Grant", mock.Anything).Return(nil)

	tenant, err := service.CreateTenant(ctx, name, "admin@example.com", "password123", creatorID)

	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, name, tenant.Name)

	uid, err := uuid.Parse(tenant.ID)
	assert.NoError(t, err)
	assert.Equal(t, byte(7), byte(uid.Version()))

	repo.AssertExpectations(t)
	authzRepo.AssertExpectations(t)
}

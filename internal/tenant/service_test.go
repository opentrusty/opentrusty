package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

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

type mockAudit struct {
	mock.Mock
}

func (m *mockAudit) Log(ctx context.Context, event audit.Event) {
	m.Called(ctx, event)
}

// TestPurpose: Validates that tenant creation correctly generates IDs using UUIDv7 for temporal ordering.
// Scope: Unit Test
// Security: Traceability and unique identification of tenants
// Expected: A new tenant is created with a valid UUIDv7 ID and the provided name.
// Test Case ID: TEN-01
func TestTenant_Service_CreateTenant_UUIDv7(t *testing.T) {
	repo := new(mockRepo)
	auditLogger := new(mockAudit)
	service := NewService(repo, nil, auditLogger)

	name := "Test Tenant"
	ctx := context.Background()

	repo.On("Create", ctx, mock.MatchedBy(func(t *Tenant) bool {
		// Verify it's a valid UUID
		_, err := uuid.Parse(t.ID)
		if err != nil {
			return false
		}
		// Verify uuid version 7
		uid, _ := uuid.Parse(t.ID)
		return uid.Version() == 7 && t.Name == name
	})).Return(nil)

	tenant, err := service.CreateTenant(ctx, name)

	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, name, tenant.Name)

	uid, err := uuid.Parse(tenant.ID)
	assert.NoError(t, err)
	assert.Equal(t, byte(7), byte(uid.Version()))

	repo.AssertExpectations(t)
}

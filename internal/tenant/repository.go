package tenant

import (
	"context"
	"errors"
)

var (
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role assignment already exists")
)

// Repository defines the interface for tenant storage
type Repository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id string) (*Tenant, error)
	GetByName(ctx context.Context, name string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Tenant, error)
}

// RoleRepository defines the interface for tenant role storage
type RoleRepository interface {
	AssignRole(ctx context.Context, role *TenantUserRole) error
	RevokeRole(ctx context.Context, tenantID, userID, role string) error
	GetUserRoles(ctx context.Context, tenantID, userID string) ([]*TenantUserRole, error)
	GetTenantUsers(ctx context.Context, tenantID string) ([]*TenantUserRole, error)
}

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
	"fmt"
	"strings"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/id"
	"github.com/opentrusty/opentrusty/internal/rbac"
)

// Service provides tenant management business logic
type Service struct {
	repo        Repository
	roleRepo    RoleRepository
	authzRepo   authz.AssignmentRepository
	auditLogger audit.Logger
}

// NewService creates a new tenant service
func NewService(repo Repository, roleRepo RoleRepository, authzRepo authz.AssignmentRepository, auditLogger audit.Logger) *Service {
	return &Service{
		repo:        repo,
		roleRepo:    roleRepo,
		authzRepo:   authzRepo,
		auditLogger: auditLogger,
	}
}

// CreateTenant creates a new tenant with a system-generated UUID v7 and assigns tenant_admin role to creator
func (s *Service) CreateTenant(ctx context.Context, name string, creatorUserID string) (*Tenant, error) {
	// 1. Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidTenantName
	}
	if len(name) < 3 || len(name) > 100 {
		return nil, ErrInvalidTenantName
	}

	// 2. Check for duplicate name
	existing, err := s.repo.GetByName(ctx, name)
	if err == nil && existing != nil {
		return nil, ErrTenantAlreadyExists
	}

	// 3. Generate UUID v7 (RFC 9562)
	tenantID := id.NewUUIDv7()

	now := time.Now()
	tenant := &Tenant{
		ID:        tenantID,
		Name:      name,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 4. Create tenant (repository should handle transaction if supported)
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// 5. Auto-provision tenant_admin role for creator
	assignment := &authz.Assignment{
		ID:             id.NewUUIDv7(),
		UserID:         creatorUserID,
		RoleID:         rbac.RoleIDTenantAdmin,
		Scope:          authz.ScopeTenant,
		ScopeContextID: &tenantID,
		GrantedAt:      now,
		GrantedBy:      audit.ActorSystemBootstrap, // System-granted during creation
	}

	if err := s.authzRepo.Grant(assignment); err != nil {
		// Note: In a true transaction, we'd rollback tenant creation here
		// For MVP, we log and continue
		return nil, fmt.Errorf("failed to assign tenant admin role: %w", err)
	}

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (s *Service) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

// GetTenantByName retrieves a tenant by name
func (s *Service) GetTenantByName(ctx context.Context, name string) (*Tenant, error) {
	return s.repo.GetByName(ctx, name)
}

// ListTenants lists tenants with pagination
func (s *Service) ListTenants(ctx context.Context, limit, offset int) ([]*Tenant, error) {
	return s.repo.List(ctx, limit, offset)
}

// AssignRole assigns a role to a user in a tenant
func (s *Service) AssignRole(ctx context.Context, tenantID, userID, role string, grantedBy string) error {
	// Validate role
	if role != RoleTenantOwner && role != RoleTenantAdmin && role != RoleTenantMember {
		return fmt.Errorf("invalid role: %s", role)
	}

	r := &TenantUserRole{
		ID:        id.NewUUIDv7(),
		TenantID:  tenantID,
		UserID:    userID,
		Role:      role,
		GrantedBy: grantedBy,
	}

	if err := s.roleRepo.AssignRole(ctx, r); err != nil {
		return err
	}

	// Audit role assignment
	s.auditLogger.Log(ctx, audit.Event{
		Type:     audit.TypeRoleAssigned,
		TenantID: tenantID,
		ActorID:  grantedBy,
		Resource: role,
		Metadata: map[string]any{audit.AttrActorID: userID},
	})

	return nil
}

// RevokeRole revokes a role from a user in a tenant
func (s *Service) RevokeRole(ctx context.Context, tenantID, userID, role string) error {
	if err := s.roleRepo.RevokeRole(ctx, tenantID, userID, role); err != nil {
		return err
	}

	// Audit role revocation (Note: ActorID logic needs context, here assumed context or empty.
	// We might need to pass `revokedBy` similar to `grantedBy` but for now we'll rely on ActorID if context provided it, or leave empty)
	// Actually, `ctx` doesn't inherently carry ActorID unless we standardise it.
	// For now, let's leave ActorID empty or "system" if unknown.
	s.auditLogger.Log(ctx, audit.Event{
		Type:     audit.TypeRoleRevoked,
		TenantID: tenantID,
		Resource: role,
		Metadata: map[string]any{audit.AttrActorID: userID},
	})

	return nil
}

// GetUserRoles retrieves all roles a user has in a tenant
func (s *Service) GetUserRoles(ctx context.Context, tenantID, userID string) ([]*TenantUserRole, error) {
	return s.roleRepo.GetUserRoles(ctx, tenantID, userID)
}

// GetTenantUsers retrieves all users with roles in a tenant
func (s *Service) GetTenantUsers(ctx context.Context, tenantID string) ([]*TenantUserRole, error) {
	return s.roleRepo.GetTenantUsers(ctx, tenantID)
}

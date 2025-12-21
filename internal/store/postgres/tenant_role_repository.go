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

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/opentrusty/opentrusty/internal/tenant"
)

// TenantRoleRepository implements tenant.RoleRepository
type TenantRoleRepository struct {
	db *DB
}

// NewTenantRoleRepository creates a new tenant role repository
func NewTenantRoleRepository(db *DB) *TenantRoleRepository {
	return &TenantRoleRepository{db: db}
}

// MapTenantRole maps internal tenant role names to seeded RBAC role IDs
func MapTenantRole(role string) string {
	switch role {
	case tenant.RoleTenantOwner:
		return "role:tenant:admin" // Map owner to admin for now or add role:tenant:owner
	case tenant.RoleTenantAdmin:
		return "role:tenant:admin"
	case tenant.RoleTenantMember:
		return "role:tenant:member"
	default:
		return "role:tenant:member"
	}
}

// AssignRole assigns a role to a user in a tenant
func (r *TenantRoleRepository) AssignRole(ctx context.Context, role *tenant.TenantUserRole) error {
	role.GrantedAt = time.Now()

	roleID := MapTenantRole(role.Role)

	var grantedBy sql.NullString
	if role.GrantedBy != "" {
		grantedBy = sql.NullString{String: role.GrantedBy, Valid: true}
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO rbac_assignments (id, user_id, role_id, scope, scope_context_id, granted_at, granted_by)
		VALUES ($1, $2, $3, 'tenant', $4, $5, $6)
		ON CONFLICT (user_id, role_id, scope, scope_context_id) DO NOTHING
	`, role.ID, role.UserID, roleID, role.TenantID, role.GrantedAt, grantedBy)

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RevokeRole revokes a role from a user in a tenant
func (r *TenantRoleRepository) RevokeRole(ctx context.Context, tenantID, userID, role string) error {
	roleID := MapTenantRole(role)
	result, err := r.db.pool.Exec(ctx, `
		DELETE FROM rbac_assignments
		WHERE user_id = $1 AND role_id = $2 AND scope = 'tenant' AND scope_context_id = $3
	`, userID, roleID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return tenant.ErrRoleNotFound
	}

	return nil
}

// GetUserRoles retrieves all roles a user has in a tenant
func (r *TenantRoleRepository) GetUserRoles(ctx context.Context, tenantID, userID string) ([]*tenant.TenantUserRole, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT a.id, a.scope_context_id, a.user_id, r.name, a.granted_at, a.granted_by
		FROM rbac_assignments a
		JOIN rbac_roles r ON a.role_id = r.id
		WHERE a.user_id = $1 AND a.scope = 'tenant' AND a.scope_context_id = $2
	`, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []*tenant.TenantUserRole
	for rows.Next() {
		var role tenant.TenantUserRole
		var grantedBy sql.NullString
		if err := rows.Scan(&role.ID, &role.TenantID, &role.UserID, &role.Role, &role.GrantedAt, &grantedBy); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		if grantedBy.Valid {
			role.GrantedBy = grantedBy.String
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

// GetTenantUsers retrieves all users with roles in a tenant
func (r *TenantRoleRepository) GetTenantUsers(ctx context.Context, tenantID string) ([]*tenant.TenantUserRole, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT a.id, a.scope_context_id, a.user_id, r.name, a.granted_at, a.granted_by
		FROM rbac_assignments a
		JOIN rbac_roles r ON a.role_id = r.id
		WHERE a.scope = 'tenant' AND a.scope_context_id = $1
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant users: %w", err)
	}
	defer rows.Close()

	var roles []*tenant.TenantUserRole
	for rows.Next() {
		var role tenant.TenantUserRole
		var grantedBy sql.NullString
		if err := rows.Scan(&role.ID, &role.TenantID, &role.UserID, &role.Role, &role.GrantedAt, &grantedBy); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		if grantedBy.Valid {
			role.GrantedBy = grantedBy.String
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

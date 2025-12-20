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

// AssignRole assigns a role to a user in a tenant
func (r *TenantRoleRepository) AssignRole(ctx context.Context, role *tenant.TenantUserRole) error {
	role.GrantedAt = time.Now()

	var grantedBy sql.NullString
	if role.GrantedBy != "" {
		grantedBy = sql.NullString{String: role.GrantedBy, Valid: true}
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO tenant_user_roles (id, tenant_id, user_id, role, granted_at, granted_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, role.ID, role.TenantID, role.UserID, role.Role, role.GrantedAt, grantedBy)

	if err != nil {
		// TODO: Check for unique violation return ErrRoleAlreadyExists
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RevokeRole revokes a role from a user in a tenant
func (r *TenantRoleRepository) RevokeRole(ctx context.Context, tenantID, userID, role string) error {
	result, err := r.db.pool.Exec(ctx, `
		DELETE FROM tenant_user_roles
		WHERE tenant_id = $1 AND user_id = $2 AND role = $3
	`, tenantID, userID, role)

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
		SELECT id, tenant_id, user_id, role, granted_at, granted_by
		FROM tenant_user_roles
		WHERE tenant_id = $1 AND user_id = $2
	`, tenantID, userID)
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
		SELECT id, tenant_id, user_id, role, granted_at, granted_by
		FROM tenant_user_roles
		WHERE tenant_id = $1
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

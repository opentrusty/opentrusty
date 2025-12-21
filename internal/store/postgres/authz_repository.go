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

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty/internal/authz"
)

// ProjectRepository implements authz.ProjectRepository
type ProjectRepository struct {
	db *DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(project *authz.Project) error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO projects (
			id, name, description, owner_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		project.ID, project.Name, project.Description, project.OwnerID,
		project.CreatedAt, project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(id string) (*authz.Project, error) {
	ctx := context.Background()

	var project authz.Project
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, name, description, owner_id, created_at, updated_at, deleted_at
		FROM projects
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.CreatedAt, &project.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, authz.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if deletedAt.Valid {
		project.DeletedAt = &deletedAt.Time
	}

	return &project, nil
}

// GetByName retrieves a project by name
func (r *ProjectRepository) GetByName(name string) (*authz.Project, error) {
	ctx := context.Background()

	var project authz.Project
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, name, description, owner_id, created_at, updated_at, deleted_at
		FROM projects
		WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.CreatedAt, &project.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, authz.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if deletedAt.Valid {
		project.DeletedAt = &deletedAt.Time
	}

	return &project, nil
}

// Update updates project information
func (r *ProjectRepository) Update(project *authz.Project) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE projects SET
			name = $2,
			description = $3
		WHERE id = $1 AND deleted_at IS NULL
	`,
		project.ID, project.Name, project.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return authz.ErrProjectNotFound
	}

	return nil
}

// Delete soft-deletes a project
func (r *ProjectRepository) Delete(id string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE projects SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, id, time.Now())

	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return authz.ErrProjectNotFound
	}

	return nil
}

// ListByOwner retrieves all projects owned by a user
func (r *ProjectRepository) ListByOwner(ownerID string) ([]*authz.Project, error) {
	ctx := context.Background()

	rows, err := r.db.pool.Query(ctx, `
		SELECT id, name, description, owner_id, created_at, updated_at, deleted_at
		FROM projects
		WHERE owner_id = $1 AND deleted_at IS NULL
	`, ownerID)

	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*authz.Project

	for rows.Next() {
		var project authz.Project
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.OwnerID,
			&project.CreatedAt, &project.UpdatedAt, &deletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		if deletedAt.Valid {
			project.DeletedAt = &deletedAt.Time
		}

		projects = append(projects, &project)
	}

	return projects, nil
}

// ListByUser retrieves all projects a user has access to
func (r *ProjectRepository) ListByUser(userID string) ([]*authz.Project, error) {
	ctx := context.Background()

	rows, err := r.db.pool.Query(ctx, `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at, p.updated_at, p.deleted_at
		FROM projects p
		INNER JOIN user_project_roles upr ON p.id = upr.project_id
		WHERE upr.user_id = $1 AND p.deleted_at IS NULL
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list user projects: %w", err)
	}
	defer rows.Close()

	var projects []*authz.Project

	for rows.Next() {
		var project authz.Project
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.OwnerID,
			&project.CreatedAt, &project.UpdatedAt, &deletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		if deletedAt.Valid {
			project.DeletedAt = &deletedAt.Time
		}

		projects = append(projects, &project)
	}

	return projects, nil
}

// RoleRepository implements authz.RoleRepository
type RoleRepository struct {
	db *DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role
func (r *RoleRepository) Create(role *authz.Role) error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO rbac_roles (
			id, name, scope, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		role.ID, role.Name, string(role.Scope), role.Description,
		role.CreatedAt, role.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(id string) (*authz.Role, error) {
	ctx := context.Background()

	var role authz.Role
	var scopeStr string

	err := r.db.pool.QueryRow(ctx, `
		SELECT r.id, r.name, r.scope, r.description, r.created_at, r.updated_at,
		       COALESCE(array_agg(p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		FROM rbac_roles r
		LEFT JOIN rbac_role_permissions rp ON r.id = rp.role_id
		LEFT JOIN rbac_permissions p ON rp.permission_id = p.id
		WHERE r.id = $1
		GROUP BY r.id, r.name, r.scope, r.description, r.created_at, r.updated_at
	`, id).Scan(
		&role.ID, &role.Name, &scopeStr, &role.Description,
		&role.CreatedAt, &role.UpdatedAt, &role.Permissions,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, authz.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	role.Scope = authz.Scope(scopeStr)
	return &role, nil
}

// GetByName retrieves a role by name and scope
func (r *RoleRepository) GetByName(name string, scope authz.Scope) (*authz.Role, error) {
	ctx := context.Background()

	var role authz.Role
	var scopeStr string

	err := r.db.pool.QueryRow(ctx, `
		SELECT r.id, r.name, r.scope, r.description, r.created_at, r.updated_at,
		       COALESCE(array_agg(p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		FROM rbac_roles r
		LEFT JOIN rbac_role_permissions rp ON r.id = rp.role_id
		LEFT JOIN rbac_permissions p ON rp.permission_id = p.id
		WHERE r.name = $1 AND r.scope = $2
		GROUP BY r.id, r.name, r.scope, r.description, r.created_at, r.updated_at
	`, name, string(scope)).Scan(
		&role.ID, &role.Name, &scopeStr, &role.Description,
		&role.CreatedAt, &role.UpdatedAt, &role.Permissions,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, authz.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	role.Scope = authz.Scope(scopeStr)
	return &role, nil
}

// Update updates role information
func (r *RoleRepository) Update(role *authz.Role) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE rbac_roles SET
			description = $2,
			updated_at = $3
		WHERE id = $1
	`,
		role.ID, role.Description, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return authz.ErrRoleNotFound
	}

	return nil
}

// Delete deletes a role
func (r *RoleRepository) Delete(id string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		DELETE FROM rbac_roles WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return authz.ErrRoleNotFound
	}

	return nil
}

// List retrieves all roles, optionally filtered by scope
func (r *RoleRepository) List(scope *authz.Scope) ([]*authz.Role, error) {
	ctx := context.Background()

	query := `
		SELECT r.id, r.name, r.scope, r.description, r.created_at, r.updated_at,
		       COALESCE(array_agg(p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		FROM rbac_roles r
		LEFT JOIN rbac_role_permissions rp ON r.id = rp.role_id
		LEFT JOIN rbac_permissions p ON rp.permission_id = p.id
	`
	var args []interface{}
	if scope != nil {
		query += " WHERE r.scope = $1"
		args = append(args, string(*scope))
	}
	query += " GROUP BY r.id, r.name, r.scope, r.description, r.created_at, r.updated_at"

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []*authz.Role

	for rows.Next() {
		var role authz.Role
		var scopeStr string

		if err := rows.Scan(
			&role.ID, &role.Name, &scopeStr, &role.Description,
			&role.CreatedAt, &role.UpdatedAt, &role.Permissions,
		); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		role.Scope = authz.Scope(scopeStr)
		roles = append(roles, &role)
	}

	return roles, nil
}

// AssignmentRepository implements authz.AssignmentRepository
type AssignmentRepository struct {
	db *DB
}

// NewAssignmentRepository creates a new assignment repository
func NewAssignmentRepository(db *DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Grant assigns a role to a user
func (r *AssignmentRepository) Grant(assignment *authz.Assignment) error {
	ctx := context.Background()

	var grantedBy sql.NullString
	if assignment.GrantedBy != "" {
		grantedBy = sql.NullString{String: assignment.GrantedBy, Valid: true}
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO rbac_assignments (
			id, user_id, role_id, scope, scope_context_id, granted_at, granted_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, role_id, scope, scope_context_id) DO NOTHING
	`,
		assignment.ID, assignment.UserID, assignment.RoleID,
		string(assignment.Scope), assignment.ScopeContextID,
		assignment.GrantedAt, grantedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to grant role: %w", err)
	}

	return nil
}

// Revoke removes a role assignment
func (r *AssignmentRepository) Revoke(userID, roleID string, scope authz.Scope, scopeContextID *string) error {
	ctx := context.Background()

	var query string
	var args []interface{}

	if scopeContextID == nil {
		query = `
			DELETE FROM rbac_assignments
			WHERE user_id = $1 AND role_id = $2 AND scope = $3 AND scope_context_id IS NULL
		`
		args = []interface{}{userID, roleID, string(scope)}
	} else {
		query = `
			DELETE FROM rbac_assignments
			WHERE user_id = $1 AND role_id = $2 AND scope = $3 AND scope_context_id = $4
		`
		args = []interface{}{userID, roleID, string(scope), *scopeContextID}
	}

	_, err := r.db.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	return nil
}

// ListForUser retrieves all assignments for a user
func (r *AssignmentRepository) ListForUser(userID string) ([]*authz.Assignment, error) {
	ctx := context.Background()

	rows, err := r.db.pool.Query(ctx, `
		SELECT id, user_id, role_id, scope, scope_context_id, granted_at, COALESCE(granted_by, '')
		FROM rbac_assignments
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list user assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*authz.Assignment

	for rows.Next() {
		var a authz.Assignment
		var scopeStr string
		var grantedBy string

		if err := rows.Scan(
			&a.ID, &a.UserID, &a.RoleID, &scopeStr, &a.ScopeContextID,
			&a.GrantedAt, &grantedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		a.Scope = authz.Scope(scopeStr)
		a.GrantedBy = grantedBy
		assignments = append(assignments, &a)
	}

	return assignments, nil
}

// ListByRole retrieves all users assigned a specific role at a scope
func (r *AssignmentRepository) ListByRole(roleID string, scope authz.Scope, scopeContextID *string) ([]string, error) {
	ctx := context.Background()

	var query string
	var args []interface{}

	if scopeContextID == nil {
		query = `
			SELECT user_id FROM rbac_assignments
			WHERE role_id = $1 AND scope = $2 AND scope_context_id IS NULL
		`
		args = []interface{}{roleID, string(scope)}
	} else {
		query = `
			SELECT user_id FROM rbac_assignments
			WHERE role_id = $1 AND scope = $2 AND scope_context_id = $3
		`
		args = []interface{}{roleID, string(scope), *scopeContextID}
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by role: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// CheckExists checks if a specific assignment exists
func (r *AssignmentRepository) CheckExists(roleID string, scope authz.Scope, scopeContextID *string) (bool, error) {
	ctx := context.Background()

	var query string
	var args []interface{}

	if scopeContextID == nil {
		query = `
			SELECT EXISTS (
				SELECT 1 FROM rbac_assignments
				WHERE role_id = $1 AND scope = $2 AND scope_context_id IS NULL
			)
		`
		args = []interface{}{roleID, string(scope)}
	} else {
		query = `
			SELECT EXISTS (
				SELECT 1 FROM rbac_assignments
				WHERE role_id = $1 AND scope = $2 AND scope_context_id = $3
			)
		`
		args = []interface{}{roleID, string(scope), *scopeContextID}
	}

	var exists bool
	err := r.db.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check assignment existence: %w", err)
	}

	return exists, nil
}

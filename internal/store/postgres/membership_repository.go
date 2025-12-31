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
	"fmt"
	"time"

	"github.com/opentrusty/opentrusty/internal/tenant"
)

// MembershipRepository implements tenant.MembershipRepository using PostgreSQL
type MembershipRepository struct {
	db *DB
}

// NewMembershipRepository creates a new PostgreSQL membership repository
func NewMembershipRepository(db *DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

// Create inserts a new membership record
func (r *MembershipRepository) Create(ctx context.Context, m *tenant.Membership) error {
	query := `
		INSERT INTO tenant_members (id, tenant_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, user_id) DO NOTHING`

	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}

	_, err := r.db.pool.Exec(ctx, query, m.ID, m.TenantID, m.UserID, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}
	return nil
}

// Delete removes a specific membership record
func (r *MembershipRepository) Delete(ctx context.Context, tenantID, userID string) error {
	query := `DELETE FROM tenant_members WHERE tenant_id = $1 AND user_id = $2`
	_, err := r.db.pool.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}
	return nil
}

// ListByUser retrieves all memberships for a user
func (r *MembershipRepository) ListByUser(ctx context.Context, userID string) ([]*tenant.Membership, error) {
	query := `
		SELECT id, tenant_id, user_id, created_at
		FROM tenant_members
		WHERE user_id = $1`

	rows, err := r.db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user memberships: %w", err)
	}
	defer rows.Close()

	var result []*tenant.Membership
	for rows.Next() {
		m := &tenant.Membership{}
		if err := rows.Scan(&m.ID, &m.TenantID, &m.UserID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		result = append(result, m)
	}
	return result, nil
}

// ListByTenant retrieves all memberships for a tenant
func (r *MembershipRepository) ListByTenant(ctx context.Context, tenantID string) ([]*tenant.Membership, error) {
	query := `
		SELECT id, tenant_id, user_id, created_at
		FROM tenant_members
		WHERE tenant_id = $1`

	rows, err := r.db.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant memberships: %w", err)
	}
	defer rows.Close()

	var result []*tenant.Membership
	for rows.Next() {
		m := &tenant.Membership{}
		if err := rows.Scan(&m.ID, &m.TenantID, &m.UserID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		result = append(result, m)
	}
	return result, nil
}

// DeleteByTenantID removes all memberships for a tenant (used during tenant deletion)
func (r *MembershipRepository) DeleteByTenantID(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenant_members WHERE tenant_id = $1`
	_, err := r.db.pool.Exec(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant memberships: %w", err)
	}
	return nil
}

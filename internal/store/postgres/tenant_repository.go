package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty/internal/tenant"
)

// TenantRepository implements tenant.Repository
type TenantRepository struct {
	db *DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create creates a new tenant
func (r *TenantRepository) Create(ctx context.Context, t *tenant.Tenant) error {
	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO tenants (id, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, t.ID, t.Name, t.Status, t.CreatedAt, t.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id string) (*tenant.Tenant, error) {
	var t tenant.Tenant
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, name, status, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&t.ID, &t.Name, &t.Status, &t.CreatedAt, &t.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, tenant.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &t, nil
}

// GetByName retrieves a tenant by name
func (r *TenantRepository) GetByName(ctx context.Context, name string) (*tenant.Tenant, error) {
	var t tenant.Tenant
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, name, status, created_at, updated_at, deleted_at
		FROM tenants
		WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(
		&t.ID, &t.Name, &t.Status, &t.CreatedAt, &t.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, tenant.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &t, nil
}

// Update updates a tenant
func (r *TenantRepository) Update(ctx context.Context, t *tenant.Tenant) error {
	t.UpdatedAt = time.Now()
	result, err := r.db.pool.Exec(ctx, `
		UPDATE tenants SET name = $2, status = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`, t.ID, t.Name, t.Status, t.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return tenant.ErrTenantNotFound
	}

	return nil
}

// Delete soft-deletes a tenant
func (r *TenantRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.pool.Exec(ctx, `
		UPDATE tenants SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, id, time.Now())

	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return tenant.ErrTenantNotFound
	}

	return nil
}

// List lists tenants
func (r *TenantRepository) List(ctx context.Context, limit, offset int) ([]*tenant.Tenant, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, name, status, created_at, updated_at
		FROM tenants
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*tenant.Tenant
	for rows.Next() {
		var t tenant.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, &t)
	}

	return tenants, nil
}

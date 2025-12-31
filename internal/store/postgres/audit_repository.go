package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/opentrusty/opentrusty/internal/audit"
)

// AuditRepository implements audit.Repository
type AuditRepository struct {
	db *DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Log persists an event
func (r *AuditRepository) Log(ctx context.Context, event audit.Event) error {
	var tenantID *string
	if event.TenantID != "" {
		tenantID = &event.TenantID
	}
	var actorID *string
	if event.ActorID != "" {
		actorID = &event.ActorID
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO audit_events (
			id, type, tenant_id, actor_id, resource, target_name, target_id, ip_address, user_agent, metadata, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`,
		event.Type,
		tenantID,
		actorID,
		event.Resource,
		event.TargetName,
		event.TargetID,
		event.IPAddress,
		event.UserAgent,
		event.Metadata,
		event.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	return nil
}

// List retrieves events matching filter
func (r *AuditRepository) List(ctx context.Context, filter audit.Filter) ([]audit.Event, int, error) {
	whereClauses := []string{}
	args := []any{}
	argIdx := 1

	if filter.TenantID != nil {
		if *filter.TenantID == "" {
			// System level?
			whereClauses = append(whereClauses, "e.tenant_id IS NULL")
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("e.tenant_id = $%d", argIdx))
			args = append(args, *filter.TenantID)
			argIdx++
		}
	}
	if filter.ActorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("e.actor_id = $%d", argIdx))
		args = append(args, *filter.ActorID)
		argIdx++
	}
	if filter.Type != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("e.type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.StartDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("e.created_at >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("e.created_at <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count Data
	countQuery := "SELECT COUNT(*) FROM audit_events e " + whereSQL
	var total int
	err := r.db.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	// Select Data
	query := `
		SELECT e.id, e.type, COALESCE(e.tenant_id, ''), COALESCE(e.actor_id, ''), 
               COALESCE(NULLIF(u.full_name, ''), NULLIF(u.email, ''), e.actor_id, ''), e.resource, 
               COALESCE(e.target_name, ''), COALESCE(e.target_id, ''), COALESCE(e.ip_address, ''), COALESCE(e.user_agent, ''), e.metadata, e.created_at
		FROM audit_events e
		LEFT JOIN users u ON e.actor_id = u.id::text
	` + whereSQL + fmt.Sprintf(" ORDER BY e.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit events: %w", err)
	}
	defer rows.Close()

	var events []audit.Event
	for rows.Next() {
		var e audit.Event

		if err := rows.Scan(
			&e.ID, &e.Type, &e.TenantID, &e.ActorID, &e.ActorName, &e.Resource,
			&e.TargetName, &e.TargetID, &e.IPAddress, &e.UserAgent, &e.Metadata, &e.Timestamp,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit event: %w", err)
		}
		events = append(events, e)
	}

	return events, total, nil
}

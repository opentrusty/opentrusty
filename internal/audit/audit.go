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

package audit

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

// Event types
const (
	TypeLoginSuccess           = "login_success"
	TypeLoginFailed            = "login_failed"
	TypeTokenIssued            = "token_issued"
	TypeTokenRevoked           = "token_revoked"
	TypeRoleAssigned           = "role_assigned"
	TypeRoleRevoked            = "role_revoked"
	TypeClientCreated          = "client_created"
	TypeSecretRotated          = "secret_rotated"
	TypeUserLocked             = "user_locked"
	TypeUserUnlocked           = "user_unlocked"
	TypeUserCreated            = "user_created"
	TypePasswordChanged        = "password_changed"
	TypeLogout                 = "logout"
	TypePlatformAdminBootstrap = "platform_admin_bootstrap"
	TypeTenantCreated          = "tenant_created"
	TypeTenantUpdated          = "tenant_updated"
	TypeTenantDeleted          = "tenant_deleted"
	TypeClientDeleted          = "client_deleted"
	TypeClientUpdated          = "client_updated"
	TypeUserUpdated            = "user_updated"
	// TypeAuditRead is emitted when a platform admin accesses tenant audit logs
	TypeAuditRead = "audit.read"
	// TypeAuditReadCrossTenant is emitted when a platform admin declares intent for cross-tenant audit access
	TypeAuditReadCrossTenant = "audit.read.cross_tenant"
)

// Standard audit attribute keys
const (
	AttrAuditType  = "audit_type"
	AttrTenantID   = "tenant_id"
	AttrActorID    = "actor_id"
	AttrActorName  = "actor_name"
	AttrResource   = "resource"
	AttrTargetName = "target_name"
	AttrTargetID   = "target_id"
	AttrTimestamp  = "timestamp"
	AttrIPAddress  = "ip_address"
	AttrUserAgent  = "user_agent"
	AttrComponent  = "component"
	AttrMetadata   = "metadata"
)

// Common Resource Types
const (
	ResourcePlatform        = "platform"
	ResourceTenant          = "tenant"
	ResourceUser            = "user"
	ResourceRole            = "role"
	ResourceClient          = "client"
	ResourceSession         = "session"
	ResourceUserCredentials = "user_credentials"
	ResourceToken           = "token"
)

// Standard Actor IDs
const (
	ActorSystemBootstrap = ""
)

// Common Metadata Keys
const (
	AttrEmail      = "email"
	AttrRoleID     = "role_id"
	AttrReason     = "reason"
	AttrAttempts   = "attempts"
	AttrSessionID  = "session_id"
	AttrTenantName = "tenant_name"
)

// Event represents an auditable action
type Event struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	TenantID   string         `json:"tenant_id"`
	ActorID    string         `json:"actor_id"`
	ActorName  string         `json:"actor_name"`
	Resource   string         `json:"resource"`
	TargetName string         `json:"target_name"`
	TargetID   string         `json:"target_id"`
	Metadata   map[string]any `json:"metadata"`
	Timestamp  time.Time      `json:"created_at"` // Match frontend expectation
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
}

// Logger defines the interface for audit logging
type Logger interface {
	Log(ctx context.Context, event Event)
}

// Filter defines criteria for listing audit events
type Filter struct {
	TenantID  *string
	ActorID   *string
	Type      *string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

// Repository defines storage for audit events
type Repository interface {
	// Log persists an event
	Log(ctx context.Context, event Event) error
	// List retrieves events matching filter
	List(ctx context.Context, filter Filter) ([]Event, int, error)
}

// SlogLogger implements Logger using slog
type SlogLogger struct{}

// NewSlogLogger creates a new audit logger
func NewSlogLogger() *SlogLogger {
	return &SlogLogger{}
}

// Log records an audit event
func (l *SlogLogger) Log(ctx context.Context, event Event) {
	// Ensure timestamp is set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Prepare attributes
	attrs := []any{
		slog.String(AttrAuditType, event.Type),
		slog.String(AttrTenantID, event.TenantID),
		slog.String(AttrActorID, event.ActorID),
		slog.String(AttrActorName, event.ActorName),
		slog.String(AttrResource, event.Resource),
		slog.String(AttrTargetName, event.TargetName),
		slog.String(AttrTargetID, event.TargetID),
		slog.Time(AttrTimestamp, event.Timestamp),
	}

	if event.IPAddress != "" {
		attrs = append(attrs, slog.String(AttrIPAddress, event.IPAddress))
	}
	if event.UserAgent != "" {
		attrs = append(attrs, slog.String(AttrUserAgent, event.UserAgent))
	}

	// Flatten metadata
	if len(event.Metadata) > 0 {
		group := []any{}
		for k, v := range event.Metadata {
			// Redact secrets
			if isSecret(k) {
				v = "[REDACTED]"
			}
			group = append(group, slog.Any(k, v))
		}
		attrs = append(attrs, slog.Group(AttrMetadata, group...))
	}

	// Log at INFO level with "audit" component
	slog.InfoContext(ctx, "AUDIT_EVENT", append(attrs, slog.String(AttrComponent, "audit"))...)
}

// RepositoryLogger implements Logger using a Repository and Slog
type RepositoryLogger struct {
	repo Repository
	slog *SlogLogger
}

// NewRepositoryLogger creates a new repository-backed logger
func NewRepositoryLogger(repo Repository) *RepositoryLogger {
	return &RepositoryLogger{
		repo: repo,
		slog: NewSlogLogger(),
	}
}

// Log records an audit event to both Slog and Repository
func (l *RepositoryLogger) Log(ctx context.Context, event Event) {
	// Ensure timestamp is set before processing
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 1. Log to Slog (Stdout)
	l.slog.Log(ctx, event)

	// 2. Persist to Repository
	// We use a detached context or error handling?
	// For now, synchronous execution to ensure audit trial integrity.
	if err := l.repo.Log(ctx, event); err != nil {
		slog.ErrorContext(ctx, "failed to persist audit event", "error", err)
	}
}

// Check if isSecret needs to be exported or not. It is used in SlogLogger, so likely private in package.
// Nothing else changed.

// isSecret checks if a key likely contains a secret.
// It uses case-insensitive substring matching against a set of common sensitive keywords.
func isSecret(key string) bool {
	// Case-insensitive check
	k := strings.ToLower(key)
	secrets := []string{
		"password", "secret", "token", "key", "authorization",
		"hash", "credential", "private", "api_key",
	}
	for _, s := range secrets {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}

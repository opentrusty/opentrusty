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
)

// Standard audit attribute keys
const (
	AttrAuditType = "audit_type"
	AttrTenantID  = "tenant_id"
	AttrActorID   = "actor_id"
	AttrResource  = "resource"
	AttrTimestamp = "timestamp"
	AttrIPAddress = "ip_address"
	AttrUserAgent = "user_agent"
	AttrComponent = "component"
	AttrMetadata  = "metadata"
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
	ActorSystemBootstrap = "system:bootstrap"
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
	Type      string
	TenantID  string
	ActorID   string
	Resource  string
	Metadata  map[string]any
	Timestamp time.Time
	IPAddress string
	UserAgent string
}

// Logger defines the interface for audit logging
type Logger interface {
	Log(ctx context.Context, event Event)
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
		slog.String(AttrResource, event.Resource),
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

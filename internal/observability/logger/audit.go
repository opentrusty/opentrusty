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

package logger

import (
	"context"
	"log/slog"
)

// AuditEvent represents a security or compliance-relevant event
type AuditEvent struct {
	EventType string
	UserID    string
	SessionID string
	ClientID  string
	IPAddress string
	Action    string
	Resource  string
	Result    string // success, failure, denied
	Reason    string
	Metadata  map[string]any
}

// AuditLogger provides methods for logging security and audit events
type AuditLogger struct {
	logger *slog.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *slog.Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger.With(Component("audit")),
	}
}

// Log logs an audit event
func (a *AuditLogger) Log(ctx context.Context, event AuditEvent) {
	attrs := []slog.Attr{
		slog.String("event_type", event.EventType),
		slog.String("action", event.Action),
		slog.String("result", event.Result),
	}

	if event.UserID != "" {
		attrs = append(attrs, slog.String("user_id", event.UserID))
	}
	if event.SessionID != "" {
		attrs = append(attrs, slog.String("session_id", event.SessionID))
	}
	if event.ClientID != "" {
		attrs = append(attrs, slog.String("client_id", event.ClientID))
	}
	if event.IPAddress != "" {
		attrs = append(attrs, slog.String("ip_address", event.IPAddress))
	}
	if event.Resource != "" {
		attrs = append(attrs, slog.String("resource", event.Resource))
	}
	if event.Reason != "" {
		attrs = append(attrs, slog.String("reason", event.Reason))
	}
	if len(event.Metadata) > 0 {
		attrs = append(attrs, slog.Any("metadata", event.Metadata))
	}

	a.logger.LogAttrs(ctx, slog.LevelInfo, "audit_event", attrs...)
}

// Authentication events
func (a *AuditLogger) LoginSuccess(ctx context.Context, userID, sessionID, ipAddr string) {
	a.Log(ctx, AuditEvent{
		EventType: "authentication",
		UserID:    userID,
		SessionID: sessionID,
		IPAddress: ipAddr,
		Action:    "login",
		Result:    "success",
	})
}

func (a *AuditLogger) LoginFailure(ctx context.Context, email, ipAddr, reason string) {
	a.Log(ctx, AuditEvent{
		EventType: "authentication",
		IPAddress: ipAddr,
		Action:    "login",
		Result:    "failure",
		Reason:    reason,
		Metadata:  map[string]any{"email": email},
	})
}

func (a *AuditLogger) Logout(ctx context.Context, userID, sessionID, ipAddr string) {
	a.Log(ctx, AuditEvent{
		EventType: "authentication",
		UserID:    userID,
		SessionID: sessionID,
		IPAddress: ipAddr,
		Action:    "logout",
		Result:    "success",
	})
}

// Authorization events
func (a *AuditLogger) AuthorizationGranted(ctx context.Context, userID, clientID, scope, ipAddr string) {
	a.Log(ctx, AuditEvent{
		EventType: "authorization",
		UserID:    userID,
		ClientID:  clientID,
		IPAddress: ipAddr,
		Action:    "authorize",
		Result:    "success",
		Metadata:  map[string]any{"scope": scope},
	})
}

func (a *AuditLogger) AuthorizationDenied(ctx context.Context, userID, clientID, reason, ipAddr string) {
	a.Log(ctx, AuditEvent{
		EventType: "authorization",
		UserID:    userID,
		ClientID:  clientID,
		IPAddress: ipAddr,
		Action:    "authorize",
		Result:    "denied",
		Reason:    reason,
	})
}

// Account management events
func (a *AuditLogger) AccountCreated(ctx context.Context, userID, createdBy, ipAddr string) {
	a.Log(ctx, AuditEvent{
		EventType: "account_management",
		UserID:    userID,
		IPAddress: ipAddr,
		Action:    "create_account",
		Result:    "success",
		Metadata:  map[string]any{"created_by": createdBy},
	})
}

func (a *AuditLogger) PasswordChanged(ctx context.Context, userID, ipAddr string) {
	a.Log(ctx, AuditEvent{
		EventType: "account_management",
		UserID:    userID,
		IPAddress: ipAddr,
		Action:    "change_password",
		Result:    "success",
	})
}

func (a *AuditLogger) AccountDeleted(ctx context.Context, userID, deletedBy, ipAddr string) {
	a.Log(ctx, AuditEvent{
		EventType: "account_management",
		UserID:    userID,
		IPAddress: ipAddr,
		Action:    "delete_account",
		Result:    "success",
		Metadata:  map[string]any{"deleted_by": deletedBy},
	})
}

// Access control events
func (a *AuditLogger) AccessDenied(ctx context.Context, userID, resource, reason, ipAddr string) {
	a.Log(ctx, AuditEvent{
		EventType: "access_control",
		UserID:    userID,
		IPAddress: ipAddr,
		Action:    "access",
		Resource:  resource,
		Result:    "denied",
		Reason:    reason,
	})
}

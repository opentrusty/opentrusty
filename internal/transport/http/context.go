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

package http

import "context"

type contextKey string

const (
	tenantIDKey  contextKey = "tenant_id"
	userIDKey    contextKey = "user_id"
	sessionIDKey contextKey = "session_id"
)

// GetUserID retrieves the authenticated User ID from context.
func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(userIDKey).(string); ok {
		return val
	}
	return ""
}

// GetTenantID retrieves the Tenant ID from context.
func GetTenantID(ctx context.Context) string {
	if val, ok := ctx.Value(tenantIDKey).(string); ok {
		return val
	}
	return ""
}

// GetSessionID retrieves the Session ID from context.
func GetSessionID(ctx context.Context) string {
	if val, ok := ctx.Value(sessionIDKey).(string); ok {
		return val
	}
	return ""
}

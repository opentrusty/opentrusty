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

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/opentrusty/opentrusty/internal/observability/logger"
)

// Platform Authorization Principles:
// 1. No tenant represents the platform
// 2. Platform authorization is expressed only via scoped roles
// 3. Tenant context must never be elevated to platform context
//
// Anti-Patterns (FORBIDDEN):
// - Magic tenant IDs (e.g., "default", "system", "platform")
// - Empty/NULL tenant_id implying platform privileges
// - Hardcoded role checks (use permission checks)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Log request start
			slog.InfoContext(r.Context(), "http_request_start",
				logger.RequestID(middleware.GetReqID(r.Context())),
				logger.Method(r.Method),
				logger.Path(r.URL.Path),
				logger.RemoteAddr(r.RemoteAddr),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				slog.InfoContext(r.Context(), "http_request_end",
					logger.RequestID(middleware.GetReqID(r.Context())),
					logger.Method(r.Method),
					logger.Path(r.URL.Path),
					logger.RemoteAddr(r.RemoteAddr),
					logger.UserAgent(r.UserAgent()),
					logger.StatusCode(ww.Status()),
					logger.Duration(time.Since(start).Milliseconds()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// TenantMiddleware is a no-op passthrough in the control plane model.
// Tenant context is derived EXCLUSIVELY from:
// - Session (AuthMiddleware sets tenant_id from session.TenantID)
// - OAuth2 client_id (OAuth2 handlers resolve client → tenant)
//
// X-Tenant-ID header is FORBIDDEN and will be rejected by AuthMiddleware if present.
// See: docs/architecture/tenant-context-resolution.md
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No tenant resolution from headers or query parameters.
		// This middleware exists for routing compatibility but performs no action.
		next.ServeHTTP(w, r)
	})
}

// RequireTenant enforces that a tenant context is present.
func RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := GetTenantID(r.Context())
		if tenantID == "" {
			respondError(w, http.StatusBadRequest, "tenant_id or X-Tenant-ID header is required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates session and adds user_id to context
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := h.getSessionFromCookie(r)
		if sessionID == "" {
			respondError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		sess, err := h.sessionService.Get(r.Context(), sessionID)
		if err != nil {
			h.clearSessionCookie(w)
			respondError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		// Namespace Isolation (Hardening Step)
		// admin-plane only accepts "admin" sessions.
		// auth-plane only accepts "auth" or "admin" sessions (admin can log into auth flows).
		if h.mode == "admin" && sess.Namespace != "admin" {
			respondError(w, http.StatusForbidden, "invalid session namespace for admin plane")
			return
		}

		// Refresh session
		if err := h.sessionService.Refresh(r.Context(), sessionID); err != nil {
			slog.ErrorContext(r.Context(), "failed to refresh session", logger.Error(err))
		}

		// Security hardening: Reject X-Tenant-ID header on authenticated requests
		// Tenant context MUST be derived exclusively from session.
		// See: docs/architecture/tenant-context-resolution.md
		if r.Header.Get("X-Tenant-ID") != "" {
			slog.WarnContext(r.Context(), "tenant header spoofing attempt detected on authenticated route",
				"session_id", slog.StringValue(sess.ID[:8]+"..."),
				"user_id", slog.StringValue(sess.UserID[:8]+"..."),
			)
			respondError(w, http.StatusBadRequest, "X-Tenant-ID header is not allowed on authenticated requests; tenant is derived from session")
			return
		}

		// Add user_id to context
		ctx := context.WithValue(r.Context(), userIDKey, sess.UserID)
		ctx = context.WithValue(ctx, sessionIDKey, sess.ID)

		// Inject session tenant as authoritative tenant context
		// Platform admins may have NULL tenant_id; tenant-scoped admins have a tenant_id.
		// Authorization privileges are derived from rbac_assignments, NOT from tenant_id presence.
		sessionTenant := ""
		if sess.TenantID != nil {
			sessionTenant = *sess.TenantID
		}
		ctx = context.WithValue(ctx, tenantIDKey, sessionTenant)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CSRFMiddleware protects against Cross-Site Request Forgery for state-changing requests.
// We enforce a custom header 'X-CSRF-Token'.
func (h *Handler) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only enforce for state-changing methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}

		// Enforce custom header for SPA and Form-based transitions
		// In a production system, this would be a dynamic token.
		// For the MVP skeleton, we enforce that the header MUST be present and non-empty.
		csrfToken := r.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			slog.WarnContext(r.Context(), "missing CSRF token header", "method", r.Method, "path", r.URL.Path)
			respondError(w, http.StatusForbidden, "CSRF protection: X-CSRF-Token header is required for state-changing operations")
			return
		}

		next.ServeHTTP(w, r)
	})
}

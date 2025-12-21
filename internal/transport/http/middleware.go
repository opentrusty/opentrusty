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

// TenantMiddleware resolves tenant identification from the request.
// It is optional at this layer; downstream handlers or RequireTenant middleware
// will enforce existence if the resource is tenant-scoped.
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tenantID string

		// 1. Check Header
		if tid := r.Header.Get("X-Tenant-ID"); tid != "" {
			tenantID = tid
		}

		// 2. Check Query Parameter
		if tenantID == "" {
			if tid := r.URL.Query().Get("tenant_id"); tid != "" {
				tenantID = tid
			}
		}

		// Inject into context if found
		if tenantID != "" {
			ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// RequireTenant enforces that a tenant context is present.
func RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := r.Context().Value("tenant_id").(string)
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

		// Refresh session
		if err := h.sessionService.Refresh(r.Context(), sessionID); err != nil {
			slog.ErrorContext(r.Context(), "failed to refresh session", logger.Error(err))
		}

		// Add user_id to context
		ctx := context.WithValue(r.Context(), "user_id", sess.UserID)
		ctx = context.WithValue(ctx, "session_id", sess.ID)

		// Authorization principles:
		// - All sessions have a tenant_id (NOT NULL in schema)
		// - Platform admin privileges are derived from rbac_assignments, NOT from tenant_id
		// - Empty tenantID DOES NOT imply platform privileges

		requestTenant, _ := r.Context().Value("tenant_id").(string)

		// Tenant isolation: session tenant must match request tenant if specified
		if requestTenant != "" && sess.TenantID != requestTenant {
			slog.WarnContext(r.Context(), "cross-tenant access attempt detected",
				"actor_id", sess.UserID,
				"session_tenant", sess.TenantID,
				"request_tenant", requestTenant,
			)
			respondError(w, http.StatusForbidden, "session does not belong to this tenant")
			return
		}

		// Use session tenant as the authoritative tenant context
		ctx = context.WithValue(ctx, "tenant_id", sess.TenantID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

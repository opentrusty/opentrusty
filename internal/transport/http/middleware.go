package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/opentrusty/opentrusty/internal/observability/logger"
)

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

// TenantMiddleware enforces explicit tenant resolution from request.
// Rejects with 400 Bad Request if missing.
// Fix for B-TENANT-01 (Rule 2.1 - Isolation at entry point)
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

		// 3. Fail-Closed behavior
		if tenantID == "" {
			respondError(w, http.StatusBadRequest, "tenant_id or X-Tenant-ID header is required")
			return
		}

		// Inject into context
		ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
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

		// Fix for B-TENANT-01 (Rule 2.3 - No Cross-Talk)
		// Verify that the session tenant matches the request tenant
		requestTenant, ok := r.Context().Value("tenant_id").(string)
		if ok && requestTenant != sess.TenantID {
			slog.WarnContext(r.Context(), "cross-tenant access attempt detected",
				"actor_id", sess.UserID,
				"session_tenant", sess.TenantID,
				"request_tenant", requestTenant,
			)
			respondError(w, http.StatusForbidden, "session does not belong to this tenant")
			return
		}

		// If quest didn't have a tenant_id settled yet (rare for protected routes but possible in flow),
		// we use the session tenant as truth but usually TenantMiddleware should have run first.
		ctx = context.WithValue(ctx, "tenant_id", sess.TenantID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

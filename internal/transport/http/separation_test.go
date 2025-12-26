package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRouter_Separation verifies that endpoints are strictly separated by mode.
func TestRouter_Separation(t *testing.T) {
	h := &Handler{
		sessionConfig: SessionConfig{CookieName: "test-session"},
	}
	rl := NewRateLimiter(100, 100)

	t.Run("Mode: Auth", func(t *testing.T) {
		r := NewRouter(h, rl, "auth")

		// 1. Should have Auth endpoints
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// Should NOT be 404. Implementation might return 400 (Body missing) or 403.
		// Detailed check: If it was 404, router didn't mount it.
		assert.NotEqual(t, http.StatusNotFound, w.Code, "/auth/login should exist in auth mode")

		// 2. Should NOT have Admin endpoints
		req = httptest.NewRequest("GET", "/api/v1/tenants", nil)
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code, "/api/v1/tenants should NOT exist in auth mode")
	})

	t.Run("Mode: Admin", func(t *testing.T) {
		r := NewRouter(h, rl, "admin")

		// 1. Should have Admin endpoints
		req := httptest.NewRequest("GET", "/api/v1/tenants", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// Should NOT be 404. likely 403/401 due to middleware.
		// Our handlers currently define middleware.
		assert.NotEqual(t, http.StatusNotFound, w.Code, "/api/v1/tenants should exist in admin mode")

		// 2. Should NOT have Auth endpoints
		req = httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code, "/api/v1/auth/login should NOT exist in admin mode")
	})
}

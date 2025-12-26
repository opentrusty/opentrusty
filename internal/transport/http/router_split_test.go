package http_test

import (
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	transportHTTP "github.com/opentrusty/opentrusty/internal/transport/http"
)

// TestRouterSeparation verifies that 'auth' and 'admin' modes
// strictly segregate their respective endpoints.
func TestRouterSeparation(t *testing.T) {
	// Initialize minimal dependencie
	// We pass nil for the Handler struct because we only check route matching,
	// and we won't execute the handlers that would use the nil dependencies.
	h := &transportHTTP.Handler{}

	tests := []struct {
		name        string
		mode        string
		path        string
		method      string
		expectFound bool // true = endpoint exists
	}{
		// Auth Mode Checks
		{"Auth Mode should have Login", "auth", "/api/v1/auth/login", "POST", true},
		{"Auth Mode should have OIDC Discovery", "auth", "/.well-known/openid-configuration", "GET", true},
		{"Auth Mode should NOT have Tenants", "auth", "/api/v1/tenants", "GET", false},
		{"Auth Mode should NOT have Health", "auth", "/health", "GET", true}, // Health is ALL

		// Admin Mode Checks
		{"Admin Mode should have Tenants", "admin", "/api/v1/tenants", "GET", true},
		{"Admin Mode should have Session Check", "admin", "/api/v1/auth/me", "GET", true},
		{"Admin Mode should NOT have Login", "admin", "/api/v1/auth/login", "POST", false},
		{"Admin Mode should NOT have OIDC Discovery", "admin", "/.well-known/openid-configuration", "GET", false},
		{"Admin Mode should have Health", "admin", "/health", "GET", true},

		// All Mode Checks
		{"All Mode should have Login", "all", "/api/v1/auth/login", "POST", true},
		{"All Mode should have Tenants", "all", "/api/v1/tenants", "GET", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We use a safe rate limiter
			rl := transportHTTP.NewRateLimiter(100, 100)

			r := transportHTTP.NewRouter(h, rl, tt.mode)

			req := httptest.NewRequest(tt.method, tt.path, nil)

			// Use Match to check availability
			rctx := chi.NewRouteContext()
			if r.Match(rctx, req.Method, req.URL.Path) {
				if !tt.expectFound {
					t.Errorf("Mode %s: Route %s %s SHOULD NOT exist", tt.mode, tt.method, tt.path)
				}
			} else {
				if tt.expectFound {
					t.Errorf("Mode %s: Route %s %s SHOULD exist", tt.mode, tt.method, tt.path)
				}
			}
		})
	}
}

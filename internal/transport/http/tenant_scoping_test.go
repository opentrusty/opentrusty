package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/oauth2"
	"github.com/opentrusty/opentrusty/internal/tenant"
)

func TestMetrics_UserCount(t *testing.T) {
	t.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	// Setup Repos
	mockTenantRepo := &stubTenantMetricsRepo{
		tenants: map[string]*tenant.Tenant{
			"t1": {ID: "t1", Name: "Tenant 1"},
		},
		users: map[string][]*tenant.TenantUserRole{
			"t1": {
				{UserID: "u1", Role: "owner"},
				{UserID: "u2", Role: "admin"},
			},
		},
	}

	t1ID := "t1"
	assignmentRepo := &stubAssignmentRepo{
		assignments: map[string]*authz.Assignment{
			"u1-t1": {UserID: "u1", RoleID: "owner", Scope: authz.ScopeTenant, ScopeContextID: &t1ID},
		},
	}
	roleRepo := &stubRoleRepo{
		roles: map[string]*authz.Role{
			"owner": {ID: "owner", Permissions: []string{authz.PermTenantView}},
		},
	}

	h := &Handler{
		tenantService: tenant.NewService(mockTenantRepo, mockTenantRepo, assignmentRepo, nil, nil, nil, audit.NewSlogLogger()),
		authzService:  authz.NewService(nil, roleRepo, assignmentRepo),
		oauth2Service: oauth2.NewService(&stubClientRepo{
			clients: map[string]*oauth2.Client{
				"c1": {ID: "c1", TenantID: "t1"},
			},
		}, nil, nil, nil, audit.NewSlogLogger(), nil, 0, 0, 0),
		auditRepo: &stubAuditRepo{},
	}

	req := httptest.NewRequest("GET", "/tenants/t1/metrics", nil)
	ctx := context.WithValue(req.Context(), tenantIDKey, "t1")
	ctx = context.WithValue(ctx, userIDKey, "u1")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("tenantID", "t1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.GetTenantMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var metrics tenant.TenantMetrics
	if err := json.Unmarshal(w.Body.Bytes(), &metrics); err != nil {
		t.Fatal(err)
	}

	if metrics.TotalUsers != 2 {
		t.Errorf("expected 2 users, got %d", metrics.TotalUsers)
	}
	if metrics.TotalClients != 1 {
		t.Errorf("expected 1 client, got %d", metrics.TotalClients)
	}
}

func TestAudit_SecurityBoundaries(t *testing.T) {
	t.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	h := &Handler{
		auditLogger:          audit.NewSlogLogger(),
		auditRepo:            &stubAuditRepo{},
		auditQuerySigningKey: []byte("01234567890123456789012345678901"),
	}

	t.Run("TenantOwner_AccessGranted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tenants/t1/audit", nil)
		ctx := context.WithValue(req.Context(), tenantIDKey, "t1")
		ctx = context.WithValue(ctx, userIDKey, "u1")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tenantID", "t1")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

		req = req.WithContext(ctx)

		t1ID := "t1"
		assignmentRepo := &stubAssignmentRepo{
			assignments: map[string]*authz.Assignment{
				"u1-t1": {UserID: "u1", RoleID: "owner", Scope: authz.ScopeTenant, ScopeContextID: &t1ID},
			},
		}
		roleRepo := &stubRoleRepo{
			roles: map[string]*authz.Role{
				"owner": {ID: "owner", Permissions: []string{authz.PermTenantViewAudit, authz.PermTenantView}},
			},
		}
		h.authzService = authz.NewService(nil, roleRepo, assignmentRepo)

		w := httptest.NewRecorder()
		h.ListTenantAuditEvents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("PlatformAdmin_DirectAccessDenied", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tenants/t1/audit", nil)
		ctx := context.WithValue(req.Context(), tenantIDKey, "t1")
		ctx = context.WithValue(ctx, userIDKey, "admin")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tenantID", "t1")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

		req = req.WithContext(ctx)

		assignmentRepo := &stubAssignmentRepo{
			assignments: map[string]*authz.Assignment{
				"admin-p": {UserID: "admin", RoleID: "platform_admin", Scope: authz.ScopePlatform},
			},
		}
		roleRepo := &stubRoleRepo{
			roles: map[string]*authz.Role{
				"platform_admin": {ID: "platform_admin", Permissions: []string{authz.PermPlatformManageTenants}},
			},
		}
		h.authzService = authz.NewService(nil, roleRepo, assignmentRepo)

		w := httptest.NewRecorder()
		h.ListTenantAuditEvents(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})
}

func TestOAuthClient_Scoping_Integration(t *testing.T) {
	t.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	mockClientRepo := &stubClientRepo{
		clients: make(map[string]*oauth2.Client),
	}
	h := &Handler{
		oauth2Service: oauth2.NewService(mockClientRepo, nil, nil, nil, audit.NewSlogLogger(), nil, 0, 0, 0),
		authzService: authz.NewService(nil, &stubRoleRepo{
			roles: map[string]*authz.Role{
				"admin": {ID: "admin", Permissions: []string{authz.PermTenantManageClients}},
			},
		}, &stubAssignmentRepo{
			assignments: map[string]*authz.Assignment{
				"u1-t1": {UserID: "u1", RoleID: "admin", Scope: authz.ScopeTenant, ScopeContextID: toStrPtr("t1")},
			},
		}),
		auditLogger: audit.NewSlogLogger(),
	}

	// 1. Create client in T1
	req := httptest.NewRequest("POST", "/tenants/t1/clients", jsonReader(map[string]any{
		"client_name": "App 1",
	}))
	ctx := context.WithValue(req.Context(), tenantIDKey, "t1")
	ctx = context.WithValue(ctx, userIDKey, "u1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("tenantID", "t1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.RegisterClient(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create failed: %s", w.Body.String())
	}

	var resp RegisterClientResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	clientID := resp.Client.ID

	// 2. Try to view client from T2 (Spoofed request if handler didn't check)
	req2 := httptest.NewRequest("GET", "/tenants/t2/clients/"+clientID, nil)
	ctx2 := context.WithValue(req2.Context(), tenantIDKey, "t2")
	ctx2 = context.WithValue(ctx2, userIDKey, "u1") // u1 is T1 admin, but let's say we have T2 admin u2
	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("tenantID", "t2")
	rctx2.URLParams.Add("clientID", clientID)
	ctx2 = context.WithValue(ctx2, chi.RouteCtxKey, rctx2)
	req2 = req2.WithContext(ctx2)

	// Update permissions for T2
	h.authzService = authz.NewService(nil, &stubRoleRepo{
		roles: map[string]*authz.Role{
			"admin": {ID: "admin", Permissions: []string{authz.PermTenantManageClients}},
		},
	}, &stubAssignmentRepo{
		assignments: map[string]*authz.Assignment{
			"u2-t2": {UserID: "u2", RoleID: "admin", Scope: authz.ScopeTenant, ScopeContextID: toStrPtr("t2")},
		},
	})
	ctx2 = context.WithValue(ctx2, userIDKey, "u2")
	req2 = req2.WithContext(ctx2)

	w2 := httptest.NewRecorder()
	h.GetClient(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 for cross-tenant client access, got %d", w2.Code)
	}
}

func jsonReader(v any) *jsonReaderCloser {
	b, _ := json.Marshal(v)
	return &jsonReaderCloser{json.NewEncoder(nil), b, 0}
}

type jsonReaderCloser struct {
	*json.Encoder
	data []byte
	off  int
}

func (r *jsonReaderCloser) Read(p []byte) (n int, err error) {
	if r.off >= len(r.data) {
		return 0, nil
	}
	n = copy(p, r.data[r.off:])
	r.off += n
	return n, nil
}
func (r *jsonReaderCloser) Close() error { return nil }

func toStrPtr(s string) *string { return &s }

// stubTenantMetricsRepo for metrics test
type stubTenantMetricsRepo struct {
	tenants map[string]*tenant.Tenant
	users   map[string][]*tenant.TenantUserRole
}

func (r *stubTenantMetricsRepo) Create(ctx context.Context, t *tenant.Tenant) error { return nil }
func (r *stubTenantMetricsRepo) GetByID(ctx context.Context, id string) (*tenant.Tenant, error) {
	if t, ok := r.tenants[id]; ok {
		return t, nil
	}
	return nil, tenant.ErrTenantNotFound
}
func (r *stubTenantMetricsRepo) GetByName(ctx context.Context, name string) (*tenant.Tenant, error) {
	return nil, nil
}
func (r *stubTenantMetricsRepo) List(ctx context.Context, limit, offset int) ([]*tenant.Tenant, error) {
	return nil, nil
}
func (r *stubTenantMetricsRepo) Update(ctx context.Context, t *tenant.Tenant) error { return nil }
func (r *stubTenantMetricsRepo) Delete(ctx context.Context, id string) error        { return nil }
func (r *stubTenantMetricsRepo) DeleteByTenantID(ctx context.Context, tenantID string) error {
	return nil
}

func (r *stubTenantMetricsRepo) AssignRole(ctx context.Context, role *tenant.TenantUserRole) error {
	return nil
}
func (r *stubTenantMetricsRepo) RevokeRole(ctx context.Context, tenantID, userID, role string) error {
	return nil
}
func (r *stubTenantMetricsRepo) GetUserRoles(ctx context.Context, tenantID, userID string) ([]*tenant.TenantUserRole, error) {
	return nil, nil
}
func (r *stubTenantMetricsRepo) GetTenantUsers(ctx context.Context, tenantID string) ([]*tenant.TenantUserRole, error) {
	return r.users[tenantID], nil
}

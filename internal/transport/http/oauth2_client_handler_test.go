package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/oauth2"
)

// TestListClients_Integration tests the client listing with proper tenant scoping
func TestListClients_Integration(t *testing.T) {
	// Set required encryption key for OAuth2 service
	os.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	defer os.Unsetenv("OPENID_KEY_ENCRYPTION_KEY")

	// Setup repositories
	mockClientRepo := &stubClientRepo{
		clients: map[string]*oauth2.Client{
			"c1": {ID: "c1", ClientName: "Client 1", TenantID: "t1", ClientID: "cid1"},
			"c2": {ID: "c2", ClientName: "Client 2", TenantID: "t1", ClientID: "cid2"},
			"c3": {ID: "c3", ClientName: "Client 3", TenantID: "t2", ClientID: "cid3"},
		},
	}

	// Create authorization service (we need the real one for this integration test)
	// For a pure unit test, we'd mock it, but this validates the full flow
	assignmentRepo := &stubAssignmentRepo{assignments: make(map[string]*authz.Assignment)}
	roleRepo := &stubRoleRepo{roles: make(map[string]*authz.Role)}

	// Create tenant admin role with manage_clients permission
	adminRole := &authz.Role{
		ID:          "admin-role",
		Name:        "tenant_admin",
		Scope:       authz.ScopeTenant,
		Permissions: []string{authz.PermTenantManageClients},
	}
	roleRepo.roles["admin-role"] = adminRole

	// Assign the role to user u1 for tenant t1
	tenantID := "t1"
	assignmentRepo.assignments["u1-admin-t1"] = &authz.Assignment{
		ID:             "u1-admin-t1",
		UserID:         "u1",
		RoleID:         "admin-role",
		Scope:          authz.ScopeTenant,
		ScopeContextID: &tenantID,
	}

	authzSvc := authz.NewService(nil, roleRepo, assignmentRepo)
	oauth2Svc := oauth2.NewService(mockClientRepo, nil, nil, nil, audit.NewSlogLogger(), nil, 0, 0, 0)

	h := &Handler{
		oauth2Service: oauth2Svc,
		authzService:  authzSvc,
		auditLogger:   audit.NewSlogLogger(),
	}

	// Test 1: Valid request with proper permissions
	req := httptest.NewRequest("GET", "/tenants/t1/clients", nil)
	ctx := context.WithValue(req.Context(), tenantIDKey, "t1")
	ctx = context.WithValue(ctx, userIDKey, "u1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.ListClients(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Clients []interface{} `json:"clients"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Clients) != 2 {
		t.Errorf("expected 2 clients for tenant t1, got %d", len(resp.Clients))
	}

	// Test 2: Forbidden request (user without permission)
	req2 := httptest.NewRequest("GET", "/tenants/t1/clients", nil)
	ctx2 := context.WithValue(req2.Context(), tenantIDKey, "t1")
	ctx2 = context.WithValue(ctx2, userIDKey, "u-unauthorized")
	req2 = req2.WithContext(ctx2)

	w2 := httptest.NewRecorder()
	h.ListClients(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unauthorized user, got %d", w2.Code)
	}
}

// TestRegisterClient_Integration tests the client registration flow
func TestRegisterClient_Integration(t *testing.T) {
	// Set required encryption key for OAuth2 service
	os.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	defer os.Unsetenv("OPENID_KEY_ENCRYPTION_KEY")

	mockClientRepo := &stubClientRepo{clients: make(map[string]*oauth2.Client)}

	assignmentRepo := &stubAssignmentRepo{assignments: make(map[string]*authz.Assignment)}
	roleRepo := &stubRoleRepo{roles: make(map[string]*authz.Role)}

	adminRole := &authz.Role{
		ID:          "admin-role",
		Name:        "tenant_admin",
		Scope:       authz.ScopeTenant,
		Permissions: []string{authz.PermTenantManageClients},
	}
	roleRepo.roles["admin-role"] = adminRole

	tenantID := "t1"
	assignmentRepo.assignments["u1-admin-t1"] = &authz.Assignment{
		ID:             "u1-admin-t1",
		UserID:         "u1",
		RoleID:         "admin-role",
		Scope:          authz.ScopeTenant,
		ScopeContextID: &tenantID,
	}

	authzSvc := authz.NewService(nil, roleRepo, assignmentRepo)
	oauth2Svc := oauth2.NewService(mockClientRepo, nil, nil, nil, audit.NewSlogLogger(), nil, 0, 0, 0)

	h := &Handler{
		oauth2Service: oauth2Svc,
		authzService:  authzSvc,
		auditLogger:   audit.NewSlogLogger(),
	}

	body := []byte(`{"client_name": "Test App", "redirect_uris": ["http://localhost/cb"], "allowed_scopes": ["openid"]}`)
	req := httptest.NewRequest("POST", "/tenants/t1/clients", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), tenantIDKey, "t1")
	ctx = context.WithValue(ctx, userIDKey, "u1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.RegisterClient(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d body: %s", w.Code, w.Body.String())
	}

	var resp RegisterClientResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp.ClientName != "Test App" {
		t.Errorf("expected Test App, got %s", resp.ClientName)
	}
	if resp.ClientSecret == "" {
		t.Error("expected client_secret to be returned")
	}
}

// TestDeleteClient_Integration tests the client deletion flow
func TestDeleteClient_Integration(t *testing.T) {
	// Set required encryption key for OAuth2 service
	os.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	defer os.Unsetenv("OPENID_KEY_ENCRYPTION_KEY")

	client := &oauth2.Client{ID: "c1", ClientID: "cid1", TenantID: "t1", ClientName: "Test Client"}
	mockClientRepo := &stubClientRepo{
		clients: map[string]*oauth2.Client{
			"c1": client,
		},
	}

	assignmentRepo := &stubAssignmentRepo{assignments: make(map[string]*authz.Assignment)}
	roleRepo := &stubRoleRepo{roles: make(map[string]*authz.Role)}

	adminRole := &authz.Role{
		ID:          "admin-role",
		Name:        "tenant_admin",
		Scope:       authz.ScopeTenant,
		Permissions: []string{authz.PermTenantManageClients},
	}
	roleRepo.roles["admin-role"] = adminRole

	tenantID := "t1"
	assignmentRepo.assignments["u1-admin-t1"] = &authz.Assignment{
		ID:             "u1-admin-t1",
		UserID:         "u1",
		RoleID:         "admin-role",
		Scope:          authz.ScopeTenant,
		ScopeContextID: &tenantID,
	}

	authzSvc := authz.NewService(nil, roleRepo, assignmentRepo)
	oauth2Svc := oauth2.NewService(mockClientRepo, nil, nil, nil, audit.NewSlogLogger(), nil, 0, 0, 0)

	h := &Handler{
		oauth2Service: oauth2Svc,
		authzService:  authzSvc,
		auditLogger:   audit.NewSlogLogger(),
	}

	req := httptest.NewRequest("DELETE", "/tenants/t1/clients/c1", nil)
	ctx := context.WithValue(req.Context(), tenantIDKey, "t1")
	ctx = context.WithValue(ctx, userIDKey, "u1")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("clientID", "c1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.DeleteClient(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

// Stub repositories for testing
type stubAssignmentRepo struct {
	assignments map[string]*authz.Assignment
}

func (r *stubAssignmentRepo) Grant(a *authz.Assignment) error {
	r.assignments[a.ID] = a
	return nil
}

func (r *stubAssignmentRepo) Revoke(userID, roleID string, scope authz.Scope, scopeContextID *string) error {
	return nil
}

func (r *stubAssignmentRepo) ListForUser(userID string) ([]*authz.Assignment, error) {
	var result []*authz.Assignment
	for _, a := range r.assignments {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (r *stubAssignmentRepo) ListByRole(roleID string, scope authz.Scope, scopeContextID *string) ([]string, error) {
	return nil, nil
}

func (r *stubAssignmentRepo) CheckExists(roleID string, scope authz.Scope, scopeContextID *string) (bool, error) {
	return false, nil
}

type stubRoleRepo struct {
	roles map[string]*authz.Role
}

func (r *stubRoleRepo) Create(role *authz.Role) error {
	r.roles[role.ID] = role
	return nil
}

func (r *stubRoleRepo) GetByID(id string) (*authz.Role, error) {
	if role, ok := r.roles[id]; ok {
		return role, nil
	}
	return nil, authz.ErrRoleNotFound
}

func (r *stubRoleRepo) GetByName(name string, scope authz.Scope) (*authz.Role, error) {
	for _, role := range r.roles {
		if role.Name == name && role.Scope == scope {
			return role, nil
		}
	}
	return nil, authz.ErrRoleNotFound
}

func (r *stubRoleRepo) Update(role *authz.Role) error {
	r.roles[role.ID] = role
	return nil
}

func (r *stubRoleRepo) Delete(id string) error {
	delete(r.roles, id)
	return nil
}

func (r *stubRoleRepo) List(scope *authz.Scope) ([]*authz.Role, error) {
	var result []*authz.Role
	for _, role := range r.roles {
		if scope == nil || role.Scope == *scope {
			result = append(result, role)
		}
	}
	return result, nil
}

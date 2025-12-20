package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/oauth2"
	"github.com/opentrusty/opentrusty/internal/oidc"
	"github.com/opentrusty/opentrusty/internal/session"
)

func TestProtocol_Discovery(t *testing.T) {
	// Setup OIDC service
	issuer := "https://auth.opentrusty.org"
	oidcService, _ := oidc.NewService(issuer)

	h := &Handler{
		oidcService: oidcService,
		auditLogger: audit.NewSlogLogger(),
	}

	// Create request
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()

	// Execute
	h.Discovery(w, req)

	// Verify
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var meta oidc.DiscoveryMetadata
	if err := json.Unmarshal(w.Body.Bytes(), &meta); err != nil {
		t.Fatalf("failed to unmarshal discovery metadata: %v", err)
	}

	if meta.Issuer != issuer {
		t.Errorf("expected issuer %s, got %s", issuer, meta.Issuer)
	}
}

func TestProtocol_JWKS(t *testing.T) {
	oidcService, _ := oidc.NewService("http://localhost")
	h := &Handler{
		oidcService: oidcService,
	}

	req := httptest.NewRequest("GET", "/jwks.json", nil)
	w := httptest.NewRecorder()

	h.JWKS(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var jwks oidc.JWKS
	if err := json.Unmarshal(w.Body.Bytes(), &jwks); err != nil {
		t.Fatalf("failed to unmarshal JWKS: %v", err)
	}

	if len(jwks.Keys) == 0 {
		t.Error("expected at least one key in JWKS")
	}
}

func TestProtocol_Token_BadRequest(t *testing.T) {
	h := &Handler{
		auditLogger: audit.NewSlogLogger(),
	}

	// Request without any parameters
	req := httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Token(w, req)

	// Should return 400 or 401 depending on client auth validation logic
	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized {
		t.Errorf("expected error status, got %d", w.Code)
	}
}

// Mocks for Protocol Testing
type stubClientRepo struct {
	clients map[string]*oauth2.Client
}

func (m *stubClientRepo) GetByClientID(id string) (*oauth2.Client, error) {
	if c, ok := m.clients[id]; ok {
		return c, nil
	}
	return nil, oauth2.ErrClientNotFound
}
func (m *stubClientRepo) GetByID(id string) (*oauth2.Client, error)       { return nil, nil }
func (m *stubClientRepo) Create(c *oauth2.Client) error                   { return nil }
func (m *stubClientRepo) Update(c *oauth2.Client) error                   { return nil }
func (m *stubClientRepo) Delete(id string) error                          { return nil }
func (m *stubClientRepo) ListByOwner(id string) ([]*oauth2.Client, error) { return nil, nil }

type stubCodeRepo struct {
	codes map[string]*oauth2.AuthorizationCode
}

func (m *stubCodeRepo) Create(c *oauth2.AuthorizationCode) error { m.codes[c.Code] = c; return nil }
func (m *stubCodeRepo) GetByCode(code string) (*oauth2.AuthorizationCode, error) {
	if c, ok := m.codes[code]; ok {
		return c, nil
	}
	return nil, oauth2.ErrCodeNotFound
}
func (m *stubCodeRepo) MarkAsUsed(code string) error {
	if c, ok := m.codes[code]; ok {
		c.IsUsed = true
	}
	return nil
}
func (m *stubCodeRepo) Delete(code string) error { return nil }
func (m *stubCodeRepo) DeleteExpired() error     { return nil }

type stubAccessRepo struct{}

func (m *stubAccessRepo) Create(t *oauth2.AccessToken) error                   { return nil }
func (m *stubAccessRepo) GetByTokenHash(h string) (*oauth2.AccessToken, error) { return nil, nil }
func (m *stubAccessRepo) Revoke(h string) error                                { return nil }
func (m *stubAccessRepo) DeleteExpired() error                                 { return nil }

type stubRefreshRepo struct{}

func (m *stubRefreshRepo) Create(t *oauth2.RefreshToken) error                   { return nil }
func (m *stubRefreshRepo) GetByTokenHash(h string) (*oauth2.RefreshToken, error) { return nil, nil }
func (m *stubRefreshRepo) Revoke(h string) error                                 { return nil }
func (m *stubRefreshRepo) DeleteExpired() error                                  { return nil }

// mockSessionService is harder to mock because Service is a struct.
// But checking AuthMiddleware requires sessionService.
// We can test Cross-Tenant by mocking the sessionService behavior via a custom Handler if possible,
// or just creating a real sessionService with mock Repo.
// Session Service needs SessionRepository.

type stubSessionRepo struct {
	sessions map[string]*session.Session
}

func (m *stubSessionRepo) Create(s *session.Session) error { m.sessions[s.ID] = s; return nil }
func (m *stubSessionRepo) Get(id string) (*session.Session, error) {
	if s, ok := m.sessions[id]; ok {
		return s, nil
	}
	return nil, session.ErrSessionNotFound
}
func (m *stubSessionRepo) Update(s *session.Session) error { return nil }
func (m *stubSessionRepo) Delete(id string) error          { delete(m.sessions, id); return nil }
func (m *stubSessionRepo) DeleteExpired() error            { return nil }
func (m *stubSessionRepo) DeleteByUserID(uid string) error { return nil }

func TestProtocol_HappyPath_Flow(t *testing.T) {
	// 1. Setup Dependencies
	clientRepo := &stubClientRepo{clients: map[string]*oauth2.Client{
		"client-1": {
			ClientID:            "client-1",
			ClientSecretHash:    oauth2.HashClientSecret("secret-1"),
			RedirectURIs:        []string{"https://app.com/cb"},
			AllowedScopes:       []string{"openid", "profile"},
			IsActive:            true,
			AccessTokenLifetime: 3600,
			TenantID:            "tenant-1",
		},
	}}
	codeRepo := &stubCodeRepo{codes: make(map[string]*oauth2.AuthorizationCode)}
	oauth2Svc := oauth2.NewService(clientRepo, codeRepo, &stubAccessRepo{}, &stubRefreshRepo{}, audit.NewSlogLogger(), nil)

	// Create OIDC service for ID Token generation
	oidcSvc, _ := oidc.NewService("http://localhost")

	// We need to inject oidcSvc into oauth2Svc as provider...
	// But oauth2.NewService takes interface. oidcSvc implements it?
	// Make sure oidcSvc implements oauth2.OIDCProvider interface.
	// oauth2.OIDCProvider has GenerateIDToken. oidc.Service has GenerateIDToken.
	// Yes, signatures match.
	// However, NewService arg is explicitly `oidcProvider`.
	oauth2Svc = oauth2.NewService(clientRepo, codeRepo, &stubAccessRepo{}, &stubRefreshRepo{}, audit.NewSlogLogger(), oidcSvc)

	h := &Handler{
		oauth2Service: oauth2Svc,
		oidcService:   oidcSvc, // For discovery/jwks if needed
		auditLogger:   audit.NewSlogLogger(),
	}

	// 2. Authorize Request (Create Code)
	// Currently Authorize endpoint in Handler does session check, then calls Service.
	// We want to test Token Exchange mainly (Back-channel).
	// To test Front-channel Authorize, we need a session.
	// Let's manually create a code first to simulate the user approved it.

	ctx := context.Background()
	authReq := &oauth2.AuthorizeRequest{
		ClientID:    "client-1",
		RedirectURI: "https://app.com/cb",
		Scope:       "openid",
		State:       "state-1",
		Nonce:       "nonce-1",
	}
	code, err := oauth2Svc.CreateAuthorizationCode(ctx, authReq, "user-1")
	if err != nil {
		t.Fatalf("failed to create code: %v", err)
	}

	// 3. Token Request (Exchange Code)
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", "client-1")
	form.Set("client_secret", "secret-1")
	form.Set("code", code.Code)
	form.Set("redirect_uri", "https://app.com/cb")

	req := httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Token(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if resp["access_token"] == "" {
		t.Error("missing access_token")
	}
	if resp["id_token"] == "" {
		t.Error("missing id_token")
	}
}

func TestProtocol_CrossTenant_Negative(t *testing.T) {
	// Setup Session Service
	sessRepo := &stubSessionRepo{sessions: make(map[string]*session.Session)}
	sessSvc := session.NewService(sessRepo, 24*time.Hour, 1*time.Hour)

	// Create Session for Tenant A
	ctx := context.Background()
	sess, _ := sessSvc.Create(ctx, "user-A", "127.0.0.1", "test-agent")
	sess.TenantID = "tenant-A" // Hack: force tenant ID for test if Create doesn't set it (Create takes user ID, usually user has tenant)
	// Wait, session.Create doesn't take TenantID. It gets user?
	// Let's check session.Service.Create. Maybe it sets it or we need to update session manually.
	// Assuming for now manual update works or we mocking Get.
	sessRepo.sessions[sess.ID].TenantID = "tenant-A"

	h := NewHandler(nil, sessSvc, nil, nil, nil, nil, audit.NewSlogLogger(), SessionConfig{CookieName: "session_id"})

	// Create Router with Middleware
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(TenantMiddleware) // Parses X-Tenant-ID
		r.Use(h.AuthMiddleware) // Checks Session vs Tenant
		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	// Request with Tenant B header but Tenant A session
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-Tenant-ID", "tenant-B")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sess.ID})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for cross-tenant access, got %d", w.Code)
	}
}

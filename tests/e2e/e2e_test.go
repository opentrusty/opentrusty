//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	baseURL = getEnv("OPENTRUSTY_API_URL", "http://127.0.0.1:8080")
	apiBase = baseURL + "/api/v1"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type TestClient struct {
	httpClient *http.Client
	tenantID   string
}

func NewTestClient(tenantID string) *TestClient {
	jar, _ := cookiejar.New(nil)
	return &TestClient{
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
		tenantID: tenantID,
	}
}

func (c *TestClient) Do(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if c.tenantID != "" {
		req.Header.Set("X-Tenant-ID", c.tenantID)
	}

	return c.httpClient.Do(req)
}

func TestE2E_Workflows(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	// State shared between subtests
	var (
		e2eTenantID     string
		e2eClientID     string
		e2eClientSecret string
		e2eUserEmail    string
		e2eUserPassword string
	)

	// 1. Platform Admin Flow
	t.Run("Platform Admin Flow", func(t *testing.T) {
		// Platform admin is a user in any tenant (e.g., 'sample') with a platform-scoped role.
		// The admin privilege comes from rbac_assignments, not from a special tenant.
		client := NewTestClient("sample")

		// Register Platform Admin
		email := "admin@opentrusty.local"
		password := "password123"

		resp, err := client.Do("POST", apiBase+"/auth/register", map[string]string{
			"email":    email,
			"password": password,
		})
		require.NoError(t, err)
		t.Logf("Registration status: %d", resp.StatusCode)
		// 201 Created or 409 Conflict (if already exists)
		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict)

		// Bootstrap the admin user
		// We use 'docker exec' to run the bootstrap command on the test container
		// The container name is usually docker-opentrusty_test-1 (standard compose naming)
		cmd := exec.Command("docker", "exec", "docker-opentrusty_test-1", "./opentrusty", "bootstrap")
		cmd.Env = append(os.Environ(),
			"OT_BOOTSTRAP_ADMIN_EMAIL="+email,
			"OT_BOOTSTRAP_ADMIN_TENANT_ID=sample",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "bootstrap command failed: %s", string(out))
		t.Logf("Bootstrap output: %s", string(out))

		// Login
		resp, err = client.Do("POST", apiBase+"/auth/login", map[string]string{
			"email":    email,
			"password": password,
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Create a Tenant
		tenantID := fmt.Sprintf("tenant-%d", time.Now().Unix())
		resp, err = client.Do("POST", apiBase+"/tenants", map[string]string{
			"id":   tenantID,
			"name": "E2E Test Tenant",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Register an OAuth2 Client
		resp, err = client.Do("POST", apiBase+"/tenants/"+tenantID+"/oauth2/clients", map[string]any{
			"client_name":                "E2E Testing App",
			"redirect_uris":              []string{"http://localhost:3000/callback"},
			"allowed_scopes":             []string{"openid", "profile", "email"},
			"token_endpoint_auth_method": "client_secret_basic",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var clientData struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		err = json.NewDecoder(resp.Body).Decode(&clientData)
		require.NoError(t, err)
		assert.NotEmpty(t, clientData.ClientID)
		assert.NotEmpty(t, clientData.ClientSecret)

		t.Logf("Created Tenant: %s, ClientID: %s", tenantID, clientData.ClientID)

		// Store for next flows
		e2eTenantID = tenantID
		e2eClientID = clientData.ClientID
		e2eClientSecret = clientData.ClientSecret
	})

	// 2. Tenant Admin Flow
	t.Run("Tenant Admin Flow", func(t *testing.T) {
		require.NotEmpty(t, e2eTenantID)

		client := NewTestClient(e2eTenantID)

		// Register Tenant Admin
		adminEmail := "admin@" + e2eTenantID + ".local"
		adminPassword := "admin_pass_123"

		resp, err := client.Do("POST", apiBase+"/auth/register", map[string]string{
			"email":    adminEmail,
			"password": adminPassword,
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Login
		resp, err = client.Do("POST", apiBase+"/auth/login", map[string]string{
			"email":    adminEmail,
			"password": adminPassword,
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Create End User
		userEmail := "user@" + e2eTenantID + ".local"
		userPassword := "user_pass_123"
		resp, err = client.Do("POST", apiBase+"/tenants/"+e2eTenantID+"/users", map[string]string{
			"email":    userEmail,
			"password": userPassword,
			"role":     "tenant_member",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		t.Logf("Created End User: %s", userEmail)
		e2eUserEmail = userEmail
		e2eUserPassword = userPassword
	})

	// 3. End User OIDC Flow
	t.Run("End User OIDC Flow", func(t *testing.T) {
		require.NotEmpty(t, e2eTenantID)
		require.NotEmpty(t, e2eClientID)

		client := NewTestClient(e2eTenantID)

		// 1. Authenticate user first (to have a session for /authorize)
		resp, err := client.Do("POST", apiBase+"/auth/login", map[string]string{
			"email":    e2eUserEmail,
			"password": e2eUserPassword,
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 2. Authorize
		state := "xyz123"
		nonce := "abc456"
		authURL := fmt.Sprintf("%s/oauth2/authorize?client_id=%s&response_type=code&scope=openid+profile&redirect_uri=%s&state=%s&nonce=%s&tenant_id=%s",
			baseURL, e2eClientID, url.QueryEscape("http://localhost:3000/callback"), state, nonce, e2eTenantID)

		// Use the same client (with session cookie), but don't follow redirects to localhost:3000
		client.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		resp, err = client.httpClient.Get(authURL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)

		loc, err := resp.Location()
		require.NoError(t, err)
		finalURL := loc
		assert.Contains(t, finalURL.String(), "code=")
		assert.Contains(t, finalURL.String(), "state="+state)

		code := finalURL.Query().Get("code")
		require.NotEmpty(t, code)

		// 3. Exchange Code for Token
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("redirect_uri", "http://localhost:3000/callback")
		data.Set("client_id", e2eClientID)
		data.Set("client_secret", e2eClientSecret)

		req, _ := http.NewRequest("POST", baseURL+"/oauth2/token?tenant_id="+e2eTenantID, bytes.NewBufferString(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err = client.httpClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tokenResp struct {
			AccessToken string `json:"access_token"`
			IDToken     string `json:"id_token"`
		}
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenResp.AccessToken)
		assert.NotEmpty(t, tokenResp.IDToken)

		t.Logf("Successfully obtained OIDC tokens")

		// 4. Validate via Discovery & JWKS
		resp, err = client.httpClient.Get(baseURL + "/.well-known/openid-configuration")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var config struct {
			JWKSUri string `json:"jwks_uri"`
		}
		err = json.NewDecoder(resp.Body).Decode(&config)
		require.NoError(t, err)
		assert.NotEmpty(t, config.JWKSUri)

		resp, err = client.httpClient.Get(config.JWKSUri)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var jwks struct {
			Keys []map[string]any `json:"keys"`
		}
		err = json.NewDecoder(resp.Body).Decode(&jwks)
		require.NoError(t, err)
		assert.NotEmpty(t, jwks.Keys)

		t.Logf("Verified JWKS endpoint")
	})
}

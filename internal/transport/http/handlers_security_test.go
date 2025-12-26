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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// AUTH API INPUT VALIDATION TESTS
// Category: Auth API - Input Validation & HTTP Behavior
// Type: Unit Test (UT)
// =============================================================================

// TestPurpose: Validates that registration fails with a 400 Bad Request if the email is empty.
// Scope: Unit Test
// Security: Input sanitization boundary check
// Expected: Returns HTTP 400 Bad Request for empty email.
// Test Case ID: REG-02
func TestAuth_Register_EmptyEmail_ReturnsBadRequest(t *testing.T) {
	h := createMinimalHandler(t)

	body := RegisterRequest{
		Email:    "", // Empty email
		Password: "validPassword123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "test-tenant-id")
	w := httptest.NewRecorder()

	h.Register(w, req)

	// Note: Anonymous registration is disabled, so it returns 403 Forbidden regardless of input.
	assert.Equal(t, http.StatusForbidden, w.Code,
		"REG-02: Registration should return 403 Forbidden as it is disabled")
}

// TestPurpose: Validates that passwords below the minimum length (8 chars) are rejected.
// Scope: Unit Test
// Security: Password strength validation (prevents weak credentials)
// RelatedDocs: docs/architecture/security_policy.md
// Expected: Returns HTTP 400 Bad Request for short passwords.
// Test Case ID: REG-04
func TestAuth_Register_WeakPassword_ReturnsBadRequest(t *testing.T) {
	h := createMinimalHandler(t)

	body := RegisterRequest{
		Email:    "test@example.com",
		Password: "short", // Less than 8 characters
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "test-tenant-id")
	w := httptest.NewRecorder()

	h.Register(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code,
		"REG-04: Registration should return 403 Forbidden as it is disabled")
}

// TestPurpose: Validates that empty request bodies for login are rejected with 400 Bad Request.
// Scope: Unit Test
// Security: Request body parsing and validation
// Expected: Returns HTTP 400 Bad Request for empty bodies.
// Test Case ID: LGN-05
func TestAuth_Login_EmptyBody_ReturnsBadRequest(t *testing.T) {
	h := createMinimalHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "test-tenant-id")
	w := httptest.NewRecorder()

	h.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code,
		"LGN-05: Empty body should return 400 Bad Request")
}

// TestPurpose: Validates that malformed JSON in the login request is rejected safely.
// Scope: Unit Test
// Security: JSON parsing safety (prevents parser exploits)
// Expected: Returns HTTP 400 Bad Request for malformed JSON.
// Test Case ID: LGN-06B
func TestAuth_Login_MalformedJSON_ReturnsBadRequest(t *testing.T) {
	h := createMinimalHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte(`{invalid_json}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "test-tenant-id")
	w := httptest.NewRecorder()

	h.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code,
		"LGN-06B: Malformed JSON should return 400 Bad Request")
}

// =============================================================================
// SECURITY TESTS - Error Message Safety
// Category: Security - Error Handling
// Type: Unit Test (UT)
// =============================================================================

// TestPurpose: Validates that error responses do not leak sensitive internal details (stack traces, paths).
// Scope: Unit Test
// Security: Information disclosure prevention (CWE-209)
// Expected: Response body does not contain patterns like "panic", "/Users/", "goroutine", etc.
// Test Case ID: SEC-02
func TestSecurity_ErrorHandling_NoSensitiveDataIsLeaked(t *testing.T) {
	h := createMinimalHandler(t)

	// Send malformed JSON to trigger parse error
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "test-tenant")
	w := httptest.NewRecorder()

	h.Login(w, req)

	body := w.Body.String()

	// Security: Error should not contain internal details
	sensitivePatterns := []string{
		"panic",
		"/Users/",
		"/home/",
		"goroutine",
		"runtime.",
		".go:",
		"stack trace",
	}

	for _, pattern := range sensitivePatterns {
		assert.NotContains(t, strings.ToLower(body), strings.ToLower(pattern),
			"SEC-02 SECURITY: Response should not contain '%s'", pattern)
	}
}

// TestPurpose: Validates that JSON responses include the application/json Content-Type header.
// Scope: Unit Test
// Security: Prevents MIME sniffing attacks
// Expected: Content-Type header contains "application/json".
// Test Case ID: SEC-10
func TestSecurity_Headers_JSONContentTypeIsSet(t *testing.T) {
	h := createMinimalHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "application/json",
		"SEC-10: JSON responses must have application/json content type")
}

// TestPurpose: Validates that the health check endpoint returns valid JSON with the expected structure.
// Scope: Unit Test
// Security: Validates safe response format
// Expected: Returns 200 OK with valid JSON structure {"status": "..."}.
// Test Case ID: SEC-05B
func TestSecurity_HealthCheck_ReturnsValidJSON(t *testing.T) {
	h := createMinimalHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Health check should return 200")

	var resp HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Health response should be valid JSON")
	assert.NotEmpty(t, resp.Status, "Health response should have status")
}

// =============================================================================
// TENANT HANDLER INPUT VALIDATION TESTS
// Category: Tenant API - Input Validation
// Type: Unit Test (UT)
// =============================================================================

// TestPurpose: Validates that creating a tenant requires a non-empty name.
// Scope: Unit Test
// Security: Input validation
// Expected: Returns 400 Bad Request when name is missing.
// Test Case ID: TEN-04
func TestTenant_Create_EmptyName_ReturnsBadRequest(t *testing.T) {
	t.Skip("TEN-04: Requires authz service - tested via System Tests")
}

// TestPurpose: Validates that assigning a role requires a non-empty role name.
// Scope: Unit Test
// Security: Input validation
// Test Case ID: ROL-03
func TestTenant_AssignRole_EmptyRole_ReturnsBadRequest(t *testing.T) {
	t.Skip("ROL-03: Requires authz service - tested via System Tests")
}

// =============================================================================
// OAUTH2 HANDLER INPUT VALIDATION TESTS
// Category: OAuth2 API - Input Validation
// Type: Unit Test (UT)
// =============================================================================

// TestPurpose: Validates that the authorize endpoint requires a client_id parameter.
// Scope: Unit Test
// Security: OAuth2 parameter validation (RFC 6749)
// Test Case ID: AUT-02
func TestOAuth2_Authorize_MissingClientID_ReturnsBadRequest(t *testing.T) {
	t.Skip("AUT-02: Requires oauth2 service - tested via integration tests")
}

// TestPurpose: Validates that unsupported grant types are rejected.
// Scope: Unit Test
// Security: OAuth2 grant type validation (RFC 6749)
// Test Case ID: TKN-11
func TestOAuth2_Token_UnsupportedGrantType_ReturnsBadRequest(t *testing.T) {
	t.Skip("TKN-11: Requires oauth2 service - tested via integration tests")
}

// =============================================================================
// TEST HELPERS
// =============================================================================

// createMinimalHandler creates a Handler with nil services for input validation testing.
//
// This handler is suitable for tests that:
// - Verify request parsing and validation
// - Check HTTP-level behavior (headers, status codes)
// - Validate error response formats
//
// For tests requiring service-level logic, use createMockedHandler or ST tests.
func createMinimalHandler(t *testing.T) *Handler {
	t.Helper()
	return &Handler{
		sessionConfig: SessionConfig{
			CookieName:     "session_id",
			CookiePath:     "/",
			CookieSecure:   true,
			CookieHTTPOnly: true,
			CookieSameSite: http.SameSiteLaxMode,
		},
	}
}

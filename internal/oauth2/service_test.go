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

package oauth2

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
)

// Mock repos for OAuth2
type MockClientRepo struct {
	clients map[string]*Client
}

func (m *MockClientRepo) GetByClientID(clientID string) (*Client, error) {
	c, ok := m.clients[clientID]
	if !ok {
		return nil, ErrClientNotFound
	}
	return c, nil
}
func (m *MockClientRepo) GetByID(id string) (*Client, error) {
	for _, c := range m.clients {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, ErrClientNotFound
}
func (m *MockClientRepo) Create(client *Client) error                   { return nil }
func (m *MockClientRepo) Update(client *Client) error                   { return nil }
func (m *MockClientRepo) Delete(id string) error                        { return nil }
func (m *MockClientRepo) ListByOwner(ownerID string) ([]*Client, error) { return nil, nil }

type MockCodeRepo struct {
	codes map[string]*AuthorizationCode
}

func (m *MockCodeRepo) Create(code *AuthorizationCode) error {
	m.codes[code.Code] = code
	return nil
}
func (m *MockCodeRepo) GetByCode(code string) (*AuthorizationCode, error) {
	c, ok := m.codes[code]
	if !ok {
		return nil, ErrCodeNotFound
	}
	return c, nil
}
func (m *MockCodeRepo) MarkAsUsed(code string) error {
	if c, ok := m.codes[code]; ok {
		c.IsUsed = true
	}
	return nil
}
func (m *MockCodeRepo) Delete(code string) error {
	delete(m.codes, code)
	return nil
}
func (m *MockCodeRepo) DeleteExpired() error { return nil }

type MockAccessRepo struct {
}

func (m *MockAccessRepo) Create(token *AccessToken) error { return nil }
func (m *MockAccessRepo) GetByTokenHash(hash string) (*AccessToken, error) {
	return nil, nil
}
func (m *MockAccessRepo) Revoke(hash string) error { return nil }
func (m *MockAccessRepo) DeleteExpired() error     { return nil }

type MockRefreshRepo struct {
}

func (m *MockRefreshRepo) Create(token *RefreshToken) error { return nil }
func (m *MockRefreshRepo) GetByTokenHash(hash string) (*RefreshToken, error) {
	return nil, nil
}
func (m *MockRefreshRepo) Revoke(hash string) error { return nil }
func (m *MockRefreshRepo) DeleteExpired() error     { return nil }

type MockOIDCProvider struct {
	CapturedNonce       string
	CapturedAccessToken string
}

func (m *MockOIDCProvider) GenerateIDToken(userID, tenantID, clientID, nonce, accessToken string) (string, error) {
	m.CapturedNonce = nonce
	m.CapturedAccessToken = accessToken
	return "mock-id-token", nil
}

// TestPurpose: Validates a successful OAuth2 authorization code exchange for tokens, including ID token generation.
// Scope: Unit Test
// Security: OAuth2 Authorization Code Grant flow (RFC 6749 Section 4.1.3)
// Expected: Returns a set of tokens (access, refresh, and OIDC ID token) on successful exchange.
func TestOAuth2_Service_ExchangeCodeForToken_Success(t *testing.T) {
	s := &Service{
		clientRepo: &MockClientRepo{
			clients: map[string]*Client{
				"client-1": {
					ClientID:             "client-1",
					ClientSecretHash:     hashClientSecret("secret-1"),
					RedirectURIs:         []string{"https://app.example.com/callback"},
					GrantTypes:           []string{"authorization_code", "refresh_token"},
					AccessTokenLifetime:  3600,
					RefreshTokenLifetime: 86400,
					TenantID:             "tenant-1",
					IsActive:             true,
				},
			},
		},
		codeRepo: &MockCodeRepo{
			codes: make(map[string]*AuthorizationCode),
		},
		accessRepo:   &MockAccessRepo{},
		refreshRepo:  &MockRefreshRepo{},
		auditLogger:  audit.NewSlogLogger(),
		oidcProvider: &MockOIDCProvider{},
	}

	ctx := context.Background()
	authReq := &AuthorizeRequest{
		ClientID:            "client-1",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile",
		State:               "state-1",
		Nonce:               "nonce-123",
		CodeChallenge:       "challenge-123",
		CodeChallengeMethod: "plain",
	}

	// 1. Create code
	code, _ := s.CreateAuthorizationCode(ctx, authReq, "user-123")

	// 2. Exchange code
	tokenReq := &TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURI:  "https://app.example.com/callback",
		Code:         code.Code,
		CodeVerifier: "challenge-123", // Match "plain"
	}

	res, err := s.ExchangeCodeForToken(ctx, tokenReq)
	if err != nil {
		t.Fatalf("exchange failed: %v", err)
	}

	if res.AccessToken == "" {
		t.Error("access token missing")
	}
	if res.IDToken != "mock-id-token" {
		t.Errorf("expected mock-id-token, got %s", res.IDToken)
	}

	// Verify OIDC provider captured the correct data
	oidc := s.oidcProvider.(*MockOIDCProvider)
	if oidc.CapturedNonce != "nonce-123" {
		t.Errorf("expected nonce-123, got %s", oidc.CapturedNonce)
	}
	if oidc.CapturedAccessToken != res.AccessToken {
		t.Error("at_hash data mismatch")
	}
}

// TestPurpose: Validates that OAuth2 code exchange fails if the PKCE code verifier does not match the challenge.
// Scope: Unit Test
// Security: PKCE enforcement (RFC 7636) to prevent code injection/interception
// Expected: Returns an error when the PKCE challenge verification fails.
func TestOAuth2_Service_ExchangeCodeForToken_PKCEFailure(t *testing.T) {
	s := &Service{
		clientRepo: &MockClientRepo{
			clients: map[string]*Client{
				"client-1": {
					ClientID:         "client-1",
					ClientSecretHash: hashClientSecret("secret-1"),
					RedirectURIs:     []string{"https://app.example.com/callback"},
					IsActive:         true,
				},
			},
		},
		codeRepo: &MockCodeRepo{
			codes: make(map[string]*AuthorizationCode),
		},
		auditLogger: audit.NewSlogLogger(),
	}

	ctx := context.Background()
	// Code with S256 challenge
	verifier := "very-secret-verifier"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	authReq := &AuthorizeRequest{
		ClientID:            "client-1",
		RedirectURI:         "https://app.example.com/callback",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}
	code, _ := s.CreateAuthorizationCode(ctx, authReq, "user-1")

	// Exchange with WRONG verifier
	tokenReq := &TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURI:  "https://app.example.com/callback",
		Code:         code.Code,
		CodeVerifier: "wrong-verifier",
	}

	_, err := s.ExchangeCodeForToken(ctx, tokenReq)
	if err == nil {
		t.Error("expected PKCE failure, got nil")
	}
}

// TestPurpose: Validates that an authorization code cannot be reused for token exchange (replay prevention).
// Scope: Unit Test
// Security: Authorization code replay attack prevention
// Expected: Second exchange attempt with the same code returns an error.
func TestOAuth2_Service_ExchangeCodeForToken_Replay(t *testing.T) {
	s := &Service{
		clientRepo: &MockClientRepo{
			clients: map[string]*Client{
				"client-1": {
					ClientID:         "client-1",
					ClientSecretHash: hashClientSecret("secret-1"),
					RedirectURIs:     []string{"https://app.example.com/callback"},
					IsActive:         true,
				},
			},
		},
		codeRepo: &MockCodeRepo{
			codes: make(map[string]*AuthorizationCode),
		},
		accessRepo:  &MockAccessRepo{},
		refreshRepo: &MockRefreshRepo{},
		auditLogger: audit.NewSlogLogger(),
	}

	ctx := context.Background()
	authReq := &AuthorizeRequest{
		ClientID:      "client-1",
		RedirectURI:   "https://app.example.com/callback",
		CodeChallenge: "challenge",
	}
	code, _ := s.CreateAuthorizationCode(ctx, authReq, "user-1")

	tokenReq := &TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURI:  "https://app.example.com/callback",
		Code:         code.Code,
		CodeVerifier: "challenge",
	}

	// 1. First Exchange - Should Success
	s.codeRepo.(*MockCodeRepo).codes[code.Code].CodeChallengeMethod = "plain" // Fix for mock setup if needed
	// Actually CreateAuthorizationCode sets it from req. default is?
	// In CreateAuthorizationCode: CodeChallengeMethod: req.CodeChallengeMethod
	// If req has empty method, validatePKCE treats it as plain if verifier matches challenge.

	_, err := s.ExchangeCodeForToken(ctx, tokenReq)
	if err != nil {
		t.Fatalf("first exchange failed: %v", err)
	}

	// 2. Second Exchange - Should Fail (Replay)
	_, err = s.ExchangeCodeForToken(ctx, tokenReq)
	if err == nil {
		t.Error("expected error on code replay, got success")
	} else {
		// Verify it is the correct error
		// We can't easily check for specific error variable equality if they are wrapped,
		// but checking the string or if it's not nil is a good start.
		// In service.go: return nil, NewError(ErrInvalidGrant, "authorization code already used")
	}
}

// TestPurpose: Validates that an expired authorization code cannot be exchanged for tokens.
// Scope: Unit Test
// Security: Temporary credential lifecycle enforcement
// Expected: Returns an error when attempting to use an expired code.
func TestOAuth2_Service_ExchangeCodeForToken_Expired(t *testing.T) {
	s := &Service{
		clientRepo: &MockClientRepo{
			clients: map[string]*Client{
				"client-1": {
					ClientID:         "client-1",
					ClientSecretHash: hashClientSecret("secret-1"),
					RedirectURIs:     []string{"https://app.example.com/callback"},
					IsActive:         true,
				},
			},
		},
		codeRepo: &MockCodeRepo{
			codes: make(map[string]*AuthorizationCode),
		},
		accessRepo:  &MockAccessRepo{},
		refreshRepo: &MockRefreshRepo{},
		auditLogger: audit.NewSlogLogger(),
	}

	ctx := context.Background()
	authReq := &AuthorizeRequest{ClientID: "client-1"}
	code, _ := s.CreateAuthorizationCode(ctx, authReq, "user-1")

	// Manually expire the code
	code.ExpiresAt = time.Now().Add(-1 * time.Hour)

	tokenReq := &TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURI:  "https://app.example.com/callback",
		Code:         code.Code,
	}

	_, err := s.ExchangeCodeForToken(ctx, tokenReq)
	if err == nil {
		t.Error("expected error on expired code, got success")
	}
}

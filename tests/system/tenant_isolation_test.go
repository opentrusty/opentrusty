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

// Package system provides integration tests that run against a real PostgreSQL database.
//
// Test Execution:
//
//	INTEGRATION_TEST=true go test -v ./tests/system/...
//
// Prerequisites:
//
//	docker compose up -d postgres
//
// Test Categories:
//   - TEN-*: Tenant isolation tests
//   - AUT-*: Authorization tests
//   - OA2-*: OAuth2 flow tests
//   - OID-*: OIDC compliance tests
package system

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/id"
	"github.com/opentrusty/opentrusty/internal/identity"
	"github.com/opentrusty/opentrusty/internal/oauth2"
	"github.com/opentrusty/opentrusty/internal/oidc"
	"github.com/opentrusty/opentrusty/internal/store/postgres"
	"github.com/opentrusty/opentrusty/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDB is the shared database connection for integration tests
var testDB *postgres.DB

// TestMain sets up and tears down the test database connection
func TestMain(m *testing.M) {
	// Skip if not integration test
	if os.Getenv("INTEGRATION_TEST") != "true" {
		os.Exit(0)
	}

	// Setup database
	ctx := context.Background()
	db, err := postgres.New(ctx, postgres.Config{
		Host:         getEnvOrDefault("DB_HOST", "localhost"),
		Port:         getEnvOrDefault("DB_PORT", "5432"),
		User:         getEnvOrDefault("DB_USER", "opentrusty"),
		Password:     getEnvOrDefault("DB_PASSWORD", "opentrusty_dev_password"),
		Database:     getEnvOrDefault("DB_NAME", "opentrusty"),
		SSLMode:      "disable",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	})
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}
	testDB = db

	// Apply migrations
	if err := db.Migrate(ctx, postgres.InitialSchema); err != nil {
		// Ignore errors for already existing tables
		_ = err
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// =============================================================================
// TENANT ISOLATION TESTS
// =============================================================================

// TestPurpose: Validates cross-tenant isolation ensures users in Tenant A cannot access Tenant B resources.
// Scope: Integration Test
// Security: Multi-tenancy boundary enforcement (prevents cross-tenant access)
// Expected: Users roles in Tenant A are not visible or usable in Tenant B.
// Test Case ID: TEN-01
func TestTenant_Isolation_UserFromTenantACannotAccessTenantBResources(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration test requires database (set INTEGRATION_TEST=true)")
	}

	ctx := context.Background()

	// Setup repositories
	tenantRepo := postgres.NewTenantRepository(testDB)
	roleRepo := postgres.NewTenantRoleRepository(testDB)
	authzRepo := postgres.NewAssignmentRepository(testDB)
	auditLogger := audit.NewSlogLogger()

	tenantService := tenant.NewService(tenantRepo, roleRepo, authzRepo, auditLogger)

	// Create creator users (required for RBAC assignment constraint)
	identityRepo := postgres.NewUserRepository(testDB)
	identityService := identity.NewService(identityRepo, nil, auditLogger, 5, time.Hour)

	creatorA, err := identityService.ProvisionIdentity(ctx, "", "creator-a-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "Creator A"})
	require.NoError(t, err)

	// Create Tenant A
	tenantA, err := tenantService.CreateTenant(ctx, "Tenant A - "+id.NewUUIDv7()[:8], creatorA.ID)
	require.NoError(t, err, "TEN-01: Failed to create Tenant A")

	creatorB, err := identityService.ProvisionIdentity(ctx, "", "creator-b-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "Creator B"})
	require.NoError(t, err)

	// Create Tenant B
	tenantB, err := tenantService.CreateTenant(ctx, "Tenant B - "+id.NewUUIDv7()[:8], creatorB.ID)
	require.NoError(t, err, "TEN-01: Failed to create Tenant B")

	// Verify tenants are different
	assert.NotEqual(t, tenantA.ID, tenantB.ID,
		"TEN-01: Tenants must have unique IDs")

	// Create user and assign role in Tenant A
	user, err := identityService.ProvisionIdentity(ctx, tenantA.ID, "user-a-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "User A"})
	require.NoError(t, err)

	err = tenantService.AssignRole(ctx, tenantA.ID, user.ID, tenant.RoleTenantMember, "")
	require.NoError(t, err, "TEN-01: Failed to assign role in Tenant A")

	// Verify user has role in Tenant A
	rolesA, err := tenantService.GetUserRoles(ctx, tenantA.ID, user.ID)
	require.NoError(t, err)
	assert.Len(t, rolesA, 1,
		"TEN-01: User should have 1 role in Tenant A")

	// CRITICAL: Verify user has NO roles in Tenant B
	rolesB, err := tenantService.GetUserRoles(ctx, tenantB.ID, user.ID)
	require.NoError(t, err)
	assert.Len(t, rolesB, 0,
		"TEN-01 SECURITY: User MUST NOT have any roles in Tenant B (tenant isolation)")
}

// =============================================================================
// AUTHORIZATION TESTS
// =============================================================================

// TestPurpose: Validates that a tenant admin can manage users within their own tenant.
// Scope: Integration Test
// Security: RBAC enforcement at service layer
// Permissions: tenant_admin role
// Expected: Tenant admin can successfully assign roles in their tenant.
// Test Case ID: AUT-01
func TestAuthz_TenantAdmin_CanManageUsersInOwnTenant(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration test requires database")
	}

	ctx := context.Background()

	// Setup
	tenantRepo := postgres.NewTenantRepository(testDB)
	roleRepo := postgres.NewTenantRoleRepository(testDB)
	authzRepo := postgres.NewAssignmentRepository(testDB)
	auditLogger := audit.NewSlogLogger()

	tenantService := tenant.NewService(tenantRepo, roleRepo, authzRepo, auditLogger)

	// Create creator and tenant
	identityRepo := postgres.NewUserRepository(testDB)
	identityService := identity.NewService(identityRepo, nil, auditLogger, 5, time.Hour)
	creator, err := identityService.ProvisionIdentity(ctx, "", "admin-creator-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "Admin Creator"})
	require.NoError(t, err)

	testTenant, err := tenantService.CreateTenant(ctx, "Test Tenant - "+id.NewUUIDv7()[:8], creator.ID)
	require.NoError(t, err)

	// Create admin and member
	admin, err := identityService.ProvisionIdentity(ctx, testTenant.ID, "admin-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "Admin"})
	require.NoError(t, err)
	member, err := identityService.ProvisionIdentity(ctx, testTenant.ID, "member-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "Member"})
	require.NoError(t, err)

	// Assign admin role (granted by system)
	err = tenantService.AssignRole(ctx, testTenant.ID, admin.ID, tenant.RoleTenantAdmin, "")
	require.NoError(t, err, "AUT-01: Failed to assign admin role")

	// Admin assigns member role
	err = tenantService.AssignRole(ctx, testTenant.ID, member.ID, tenant.RoleTenantMember, admin.ID)
	assert.NoError(t, err,
		"AUT-01: Tenant admin should be able to assign roles in own tenant")

	// Verify assignment
	roles, _ := tenantService.GetUserRoles(ctx, testTenant.ID, member.ID)
	assert.Len(t, roles, 1,
		"AUT-01: Member should have assigned role")
}

// TestPurpose: Validates that invalid or malicious role names are rejected during assignment.
// Scope: Integration Test
// Security: Prevents privilege escalation via role name manipulation (e.g. SQL injection or privilege escalation)
// Expected: Returns an error when an invalid role name is used.
// Test Case ID: AUT-02
func TestAuthz_RoleAssignment_InvalidRoleNameIsRejected(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration test requires database")
	}

	ctx := context.Background()

	tenantRepo := postgres.NewTenantRepository(testDB)
	roleRepo := postgres.NewTenantRoleRepository(testDB)
	authzRepo := postgres.NewAssignmentRepository(testDB)
	auditLogger := audit.NewSlogLogger()

	tenantService := tenant.NewService(tenantRepo, roleRepo, authzRepo, auditLogger)

	identityRepo := postgres.NewUserRepository(testDB)
	identityService := identity.NewService(identityRepo, nil, auditLogger, 5, time.Hour)
	creator, err := identityService.ProvisionIdentity(ctx, "", "role-creator-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "Role Creator"})
	require.NoError(t, err)

	testTenant, err := tenantService.CreateTenant(ctx, "Role Test - "+id.NewUUIDv7()[:8], creator.ID)
	require.NoError(t, err)

	user, err := identityService.ProvisionIdentity(ctx, testTenant.ID, "invalid-role-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{})
	require.NoError(t, err)

	// Attempt to assign invalid roles
	invalidRoles := []string{
		"platform_admin",     // Platform role not valid in tenant context
		"super_admin",        // Non-existent role
		"root",               // Non-existent role
		"",                   // Empty role
		"tenant_admin; DROP", // SQL injection attempt
	}

	for _, invalidRole := range invalidRoles {
		err := tenantService.AssignRole(ctx, testTenant.ID, user.ID, invalidRole, "")
		assert.Error(t, err,
			"AUT-02 SECURITY: Invalid role '%s' should be rejected", invalidRole)
	}
}

// =============================================================================
// OAUTH2 FLOW TESTS
// =============================================================================

// TestPurpose: Validates that authorization codes cannot be used more than once (replay attack prevention).
// Scope: Integration Test
// Security: Prevents authorization code replay attacks (RFC 6749 Section 4.1.2)
// RelatedDocs: docs/architecture/oauth2.md
// Expected: Second exchange attempt fails with "code already used" error.
// Test Case ID: OA2-01
func TestOAuth2_AuthorizationCode_OneTimeUseEnforced(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration test requires database")
	}

	// Set required encryption key
	os.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	defer os.Unsetenv("OPENID_KEY_ENCRYPTION_KEY")

	ctx := context.Background()

	// Setup repositories
	tenantRepo := postgres.NewTenantRepository(testDB)
	roleRepo := postgres.NewTenantRoleRepository(testDB)
	clientRepo := postgres.NewClientRepository(testDB)
	codeRepo := postgres.NewAuthorizationCodeRepository(testDB)
	accessRepo := postgres.NewAccessTokenRepository(testDB)
	refreshRepo := postgres.NewRefreshTokenRepository(testDB)
	authzRepo := postgres.NewAssignmentRepository(testDB)
	auditLogger := audit.NewSlogLogger()

	tenantService := tenant.NewService(tenantRepo, roleRepo, authzRepo, auditLogger)

	identityRepo := postgres.NewUserRepository(testDB)
	identityService := identity.NewService(identityRepo, nil, auditLogger, 5, time.Hour)
	creator, err := identityService.ProvisionIdentity(ctx, "", "oauth-creator-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "OAuth Creator"})
	require.NoError(t, err)

	// Create tenant
	testTenant, err := tenantService.CreateTenant(ctx, "OAuth2 Test - "+id.NewUUIDv7()[:8], creator.ID)
	require.NoError(t, err)

	// Create client
	client := &oauth2.Client{
		ID:                      id.NewUUIDv7(),
		TenantID:                testTenant.ID,
		ClientID:                id.NewUUIDv7(),
		ClientName:              "Test Client",
		ClientSecretHash:        oauth2.HashClientSecret("test-secret"),
		RedirectURIs:            []string{"https://app.example.com/callback"},
		AllowedScopes:           []string{"openid"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "client_secret_basic",
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		IsActive:                true,
	}
	err = clientRepo.Create(client)
	require.NoError(t, err, "OA2-01: Failed to create client")

	// Create OIDC service
	oidcService, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	// Create OAuth2 service
	oauth2Service := oauth2.NewService(
		clientRepo, codeRepo, accessRepo, refreshRepo, auditLogger, oidcService,
		5*time.Minute, 1*time.Hour, 720*time.Hour,
	)

	// Create user
	user, err := identityService.ProvisionIdentity(ctx, testTenant.ID, "oa2-01-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{})
	require.NoError(t, err)

	// Create authorization code
	authReq := &oauth2.AuthorizeRequest{
		ClientID:            client.ClientID,
		RedirectURI:         "https://app.example.com/callback",
		ResponseType:        "code",
		Scope:               "openid",
		State:               "state-" + id.NewUUIDv7()[:8],
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "plain",
	}
	code, err := oauth2Service.CreateAuthorizationCode(ctx, authReq, user.ID)
	require.NoError(t, err, "OA2-01: Failed to create auth code")

	// First exchange - should succeed
	tokenReq := &oauth2.TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
		RedirectURI:  "https://app.example.com/callback",
		Code:         code.Code,
		CodeVerifier: "test-challenge",
	}

	_, err = oauth2Service.ExchangeCodeForToken(ctx, tokenReq)
	require.NoError(t, err, "OA2-01: First exchange should succeed")

	// CRITICAL: Second exchange - must fail (replay attack)
	_, err = oauth2Service.ExchangeCodeForToken(ctx, tokenReq)
	assert.Error(t, err,
		"OA2-01 SECURITY: Second exchange MUST fail (code replay prevention)")
}

// TestPurpose: Validates that revoking a refresh token makes it unusable for obtaining new access tokens.
// Scope: Integration Test
// Security: Ensures revoked tokens cannot be used (RFC 7009)
// RelatedDocs: docs/architecture/oauth2.md
// Expected: Refresh attempt with a revoked token fails.
// Test Case ID: OA2-02
func TestOAuth2_RefreshToken_RevocationPreventsUsage(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration test requires database")
	}

	os.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	defer os.Unsetenv("OPENID_KEY_ENCRYPTION_KEY")

	ctx := context.Background()

	// Setup
	tenantRepo := postgres.NewTenantRepository(testDB)
	roleRepo := postgres.NewTenantRoleRepository(testDB)
	clientRepo := postgres.NewClientRepository(testDB)
	codeRepo := postgres.NewAuthorizationCodeRepository(testDB)
	accessRepo := postgres.NewAccessTokenRepository(testDB)
	refreshRepo := postgres.NewRefreshTokenRepository(testDB)
	authzRepo := postgres.NewAssignmentRepository(testDB)
	auditLogger := audit.NewSlogLogger()

	tenantService := tenant.NewService(tenantRepo, roleRepo, authzRepo, auditLogger)

	identityRepo := postgres.NewUserRepository(testDB)
	identityService := identity.NewService(identityRepo, nil, auditLogger, 5, time.Hour)
	creator, err := identityService.ProvisionIdentity(ctx, "", "revoke-creator-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "Revoke Creator"})
	require.NoError(t, err)

	// Create tenant and client
	testTenant, err := tenantService.CreateTenant(ctx, "Revoke Test - "+id.NewUUIDv7()[:8], creator.ID)
	require.NoError(t, err)

	client := &oauth2.Client{
		ID:                      id.NewUUIDv7(),
		TenantID:                testTenant.ID,
		ClientID:                id.NewUUIDv7(),
		ClientName:              "Revoke Test Client",
		ClientSecretHash:        oauth2.HashClientSecret("revoke-secret"),
		RedirectURIs:            []string{"https://app.example.com/callback"},
		AllowedScopes:           []string{"openid"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "client_secret_basic",
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		IsActive:                true,
	}
	err = clientRepo.Create(client)
	require.NoError(t, err)

	oidcService, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)
	oauth2Service := oauth2.NewService(
		clientRepo, codeRepo, accessRepo, refreshRepo, auditLogger, oidcService,
		5*time.Minute, 1*time.Hour, 720*time.Hour,
	)

	// Create user
	user, err := identityService.ProvisionIdentity(ctx, testTenant.ID, "oa2-02-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{})
	require.NoError(t, err)

	// Get tokens via code exchange
	authReq := &oauth2.AuthorizeRequest{
		ClientID:      client.ClientID,
		RedirectURI:   "https://app.example.com/callback",
		ResponseType:  "code",
		Scope:         "openid",
		CodeChallenge: "revoke-challenge",
	}
	code, err := oauth2Service.CreateAuthorizationCode(ctx, authReq, user.ID)
	require.NoError(t, err)

	tokenResp, err := oauth2Service.ExchangeCodeForToken(ctx, &oauth2.TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     client.ClientID,
		ClientSecret: "revoke-secret",
		RedirectURI:  "https://app.example.com/callback",
		Code:         code.Code,
		CodeVerifier: "revoke-challenge",
	})
	require.NoError(t, err)
	require.NotEmpty(t, tokenResp.RefreshToken, "OA2-02: Must have refresh token")

	// Revoke the refresh token
	err = oauth2Service.RevokeRefreshToken(ctx, tokenResp.RefreshToken, client.ClientID)
	require.NoError(t, err, "OA2-02: Revocation should succeed")

	// CRITICAL: Attempt to use revoked token - must fail
	_, err = oauth2Service.RefreshAccessToken(ctx, &oauth2.TokenRequest{
		GrantType:    "refresh_token",
		ClientID:     client.ClientID,
		ClientSecret: "revoke-secret",
		RefreshToken: tokenResp.RefreshToken,
	})
	assert.Error(t, err,
		"OA2-02 SECURITY: Revoked refresh token MUST NOT be usable")
}

// =============================================================================
// OIDC COMPLIANCE TESTS
// =============================================================================

// TestPurpose: Validates that an id_token is only issued when the 'openid' scope is explicitly requested.
// Scope: Integration Test
// Security: Compliance with OIDC Core specification (Issue #3.1.2.1)
// Permissions: scope:openid
// Expected: id_token is present only when 'openid' is in scopes.
// Test Case ID: OID-01
func TestOIDC_TokenExchange_IDTokenOnlyWithOpenIDScope(t *testing.T) {
	if testDB == nil {
		t.Skip("Integration test requires database")
	}

	os.Setenv("OPENID_KEY_ENCRYPTION_KEY", "01234567890123456789012345678901")
	defer os.Unsetenv("OPENID_KEY_ENCRYPTION_KEY")

	ctx := context.Background()

	// Setup
	tenantRepo := postgres.NewTenantRepository(testDB)
	roleRepo := postgres.NewTenantRoleRepository(testDB)
	clientRepo := postgres.NewClientRepository(testDB)
	codeRepo := postgres.NewAuthorizationCodeRepository(testDB)
	accessRepo := postgres.NewAccessTokenRepository(testDB)
	refreshRepo := postgres.NewRefreshTokenRepository(testDB)
	authzRepo := postgres.NewAssignmentRepository(testDB)
	auditLogger := audit.NewSlogLogger()

	tenantService := tenant.NewService(tenantRepo, roleRepo, authzRepo, auditLogger)

	identityRepo := postgres.NewUserRepository(testDB)
	identityService := identity.NewService(identityRepo, nil, auditLogger, 5, time.Hour)
	creator, err := identityService.ProvisionIdentity(ctx, "", "oidc-creator-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{FullName: "OIDC Creator"})
	require.NoError(t, err)

	testTenant, err := tenantService.CreateTenant(ctx, "OIDC Test - "+id.NewUUIDv7()[:8], creator.ID)
	require.NoError(t, err)

	client := &oauth2.Client{
		ID:                      id.NewUUIDv7(),
		TenantID:                testTenant.ID,
		ClientID:                id.NewUUIDv7(),
		ClientName:              "OIDC Test Client",
		ClientSecretHash:        oauth2.HashClientSecret("oidc-secret"),
		RedirectURIs:            []string{"https://app.example.com/callback"},
		AllowedScopes:           []string{"openid", "profile", "email"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "client_secret_basic",
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		IsActive:                true,
	}
	err = clientRepo.Create(client)
	require.NoError(t, err)

	oidcService, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)
	oauth2Service := oauth2.NewService(
		clientRepo, codeRepo, accessRepo, refreshRepo, auditLogger, oidcService,
		5*time.Minute, 1*time.Hour, 720*time.Hour,
	)

	// Create user
	user, err := identityService.ProvisionIdentity(ctx, testTenant.ID, "oidc-"+id.NewUUIDv7()[:8]+"@example.com", identity.Profile{})
	require.NoError(t, err)

	// Test WITH openid scope - id_token must be present
	t.Run("WithOpenIDScope_IDTokenPresent", func(t *testing.T) {
		authReq := &oauth2.AuthorizeRequest{
			ClientID:      client.ClientID,
			RedirectURI:   "https://app.example.com/callback",
			ResponseType:  "code",
			Scope:         "openid profile", // Has openid
			CodeChallenge: "oidc-challenge-1",
		}
		code, err := oauth2Service.CreateAuthorizationCode(ctx, authReq, user.ID)
		require.NoError(t, err)

		resp, err := oauth2Service.ExchangeCodeForToken(ctx, &oauth2.TokenRequest{
			GrantType:    "authorization_code",
			ClientID:     client.ClientID,
			ClientSecret: "oidc-secret",
			RedirectURI:  "https://app.example.com/callback",
			Code:         code.Code,
			CodeVerifier: "oidc-challenge-1",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.IDToken,
			"OID-01: id_token MUST be present when openid scope is requested")
	})

	// Test WITHOUT openid scope - id_token must NOT be present
	t.Run("WithoutOpenIDScope_IDTokenAbsent", func(t *testing.T) {
		authReq := &oauth2.AuthorizeRequest{
			ClientID:      client.ClientID,
			RedirectURI:   "https://app.example.com/callback",
			ResponseType:  "code",
			Scope:         "profile", // NO openid
			CodeChallenge: "oidc-challenge-2",
		}
		code, err := oauth2Service.CreateAuthorizationCode(ctx, authReq, user.ID)
		require.NoError(t, err)

		resp, err := oauth2Service.ExchangeCodeForToken(ctx, &oauth2.TokenRequest{
			GrantType:    "authorization_code",
			ClientID:     client.ClientID,
			ClientSecret: "oidc-secret",
			RedirectURI:  "https://app.example.com/callback",
			Code:         code.Code,
			CodeVerifier: "oidc-challenge-2",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.IDToken,
			"OID-01: id_token MUST NOT be present when openid scope is not requested")
	})
}

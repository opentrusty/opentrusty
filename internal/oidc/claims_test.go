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

package oidc_test

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/opentrusty/opentrusty/internal/id"
	"github.com/opentrusty/opentrusty/internal/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPurpose: Verifies that the sub (subject) claim is deterministic: the same (tenant, user) pair always produces the same sub value.
// Scope: Unit Test
// Security: Identity Stability (prevents account fragmentation)
// Expected: sub1 == sub2 for the same user and tenant.
func TestOIDC_Claims_SubClaimStability_SameTenantAndUserProduceSameSub(t *testing.T) {
	svc, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	clientID := id.NewUUIDv7()

	// Generate two tokens with the same tenant+user
	token1, err := svc.GenerateIDToken(userID, tenantID, clientID, "", "access-token-1")
	require.NoError(t, err)

	token2, err := svc.GenerateIDToken(userID, tenantID, clientID, "", "access-token-2")
	require.NoError(t, err)

	// Parse tokens to extract sub claims
	sub1 := extractClaim(t, svc, token1, "sub")
	sub2 := extractClaim(t, svc, token2, "sub")

	assert.Equal(t, sub1, sub2, "sub claim must be stable for same tenant+user")
}

// TestPurpose: Verifies that different tenants produce different sub values, even for the same user ID (Pairwise ID-like behavior).
// Scope: Unit Test
// Security: Multi-tenant Privacy and Isolation
// Expected: subA != subB for different tenants.
func TestOIDC_Claims_SubClaimStability_DifferentTenantProducesDifferentSub(t *testing.T) {
	svc, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	tenantA := id.NewUUIDv7()
	tenantB := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	clientID := id.NewUUIDv7()

	tokenA, err := svc.GenerateIDToken(userID, tenantA, clientID, "", "access-token")
	require.NoError(t, err)

	tokenB, err := svc.GenerateIDToken(userID, tenantB, clientID, "", "access-token")
	require.NoError(t, err)

	subA := extractClaim(t, svc, tokenA, "sub")
	subB := extractClaim(t, svc, tokenB, "sub")

	assert.NotEqual(t, subA, subB, "sub claim must differ for different tenants")
}

// TestPurpose: Verifies that different users in the same tenant produce different sub values.
// Scope: Unit Test
// Security: Identity Uniqueness
// Expected: subA != subB for different users.
func TestOIDC_Claims_SubClaimStability_DifferentUserProducesDifferentSub(t *testing.T) {
	svc, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	tenantID := id.NewUUIDv7()
	userA := id.NewUUIDv7()
	userB := id.NewUUIDv7()
	clientID := id.NewUUIDv7()

	tokenA, err := svc.GenerateIDToken(userA, tenantID, clientID, "", "access-token")
	require.NoError(t, err)

	tokenB, err := svc.GenerateIDToken(userB, tenantID, clientID, "", "access-token")
	require.NoError(t, err)

	subA := extractClaim(t, svc, tokenA, "sub")
	subB := extractClaim(t, svc, tokenB, "sub")

	assert.NotEqual(t, subA, subB, "sub claim must differ for different users")
}

// TestPurpose: Verifies that the nonce claim is included in the ID token when provided (OIDC Core Section 3.1.2.1).
// Scope: Unit Test
// Security: Replay Attack Prevention
// Expected: nonce claim in token matches provided nonce.
func TestOIDC_Claims_NoncePropagation_NonceIsIncludedWhenProvided(t *testing.T) {
	svc, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	clientID := id.NewUUIDv7()
	expectedNonce := "random-nonce-12345"

	token, err := svc.GenerateIDToken(userID, tenantID, clientID, expectedNonce, "access-token")
	require.NoError(t, err)

	nonce := extractClaim(t, svc, token, "nonce")
	assert.Equal(t, expectedNonce, nonce, "nonce must be propagated to ID token")
}

// TestNoncePropagation_NonceIsOmittedWhenEmpty verifies that the nonce claim
// is NOT included when not provided.
func TestNoncePropagation_NonceIsOmittedWhenEmpty(t *testing.T) {
	svc, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	clientID := id.NewUUIDv7()

	token, err := svc.GenerateIDToken(userID, tenantID, clientID, "", "access-token")
	require.NoError(t, err)

	// Parse without validation to check claims map
	parsed, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	require.NoError(t, err)

	claims := parsed.Claims.(jwt.MapClaims)
	_, hasNonce := claims["nonce"]
	assert.False(t, hasNonce, "nonce should not be present when empty")
}

// TestPurpose: Verifies that at_hash is computed as the base64url encoding of the left-most half of the SHA-256 hash of the access token.
// Scope: Unit Test
// Security: Token Binding Verification (OIDC Core Section 3.1.3.6)
// Expected: at_hash matches the specific OIDC computation.
func TestOIDC_Claims_AtHashCorrectness_AtHashIsCorrectlyComputed(t *testing.T) {
	svc, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	clientID := id.NewUUIDv7()
	accessToken := "test-access-token-for-hash-computation"

	token, err := svc.GenerateIDToken(userID, tenantID, clientID, "", accessToken)
	require.NoError(t, err)

	// Compute expected at_hash
	atHash := sha256.Sum256([]byte(accessToken))
	leftHalf := atHash[:len(atHash)/2]
	expectedAtHash := base64.RawURLEncoding.EncodeToString(leftHalf)

	actualAtHash := extractClaim(t, svc, token, "at_hash")
	assert.Equal(t, expectedAtHash, actualAtHash, "at_hash must match OIDC spec computation")
}

// TestAtHashCorrectness_AtHashIsOmittedWhenNoAccessToken verifies that at_hash is not included
// when no access token is provided.
func TestAtHashCorrectness_AtHashIsOmittedWhenNoAccessToken(t *testing.T) {
	svc, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	tenantID := id.NewUUIDv7()
	userID := id.NewUUIDv7()
	clientID := id.NewUUIDv7()

	token, err := svc.GenerateIDToken(userID, tenantID, clientID, "", "")
	require.NoError(t, err)

	// Parse without validation to check claims map
	parsed, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	require.NoError(t, err)

	claims := parsed.Claims.(jwt.MapClaims)
	_, hasAtHash := claims["at_hash"]
	assert.False(t, hasAtHash, "at_hash should not be present when access token is empty")
}

// TestPurpose: Verifies that the iss claim matches the service issuer configuration.
// Scope: Unit Test
// Security: Trust Root Verification
// Expected: iss claim matches expected issuer.
func TestOIDC_Claims_IssuerMatch(t *testing.T) {
	expectedIssuer := "https://auth.example.com"
	svc, err := oidc.NewService(expectedIssuer)
	require.NoError(t, err)

	token, err := svc.GenerateIDToken(id.NewUUIDv7(), id.NewUUIDv7(), id.NewUUIDv7(), "", "token")
	require.NoError(t, err)

	iss := extractClaim(t, svc, token, "iss")
	assert.Equal(t, expectedIssuer, iss, "iss claim must match service issuer")
}

// TestPurpose: Verifies that the aud claim matches the client ID.
// Scope: Unit Test
// Security: Intended Audience Verification
// Expected: aud claim matches client ID.
func TestOIDC_Claims_AudienceMatch(t *testing.T) {
	svc, err := oidc.NewService("https://auth.example.com")
	require.NoError(t, err)

	clientID := id.NewUUIDv7()
	token, err := svc.GenerateIDToken(id.NewUUIDv7(), id.NewUUIDv7(), clientID, "", "token")
	require.NoError(t, err)

	aud := extractClaim(t, svc, token, "aud")
	assert.Equal(t, clientID, aud, "aud claim must match client ID")
}

// extractClaim is a helper that parses a JWT and extracts a string claim.
func extractClaim(t *testing.T, svc *oidc.Service, tokenString, claimName string) string {
	t.Helper()

	// Get the JWKS for key validation
	jwks := svc.GetJWKS()
	require.Len(t, jwks.Keys, 1, "expected 1 key in JWKS")

	// Parse without validation for claim extraction
	parsed, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	require.NoError(t, err)

	claims := parsed.Claims.(jwt.MapClaims)
	val, ok := claims[claimName]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

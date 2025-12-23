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

package oidc

import (
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// TestPurpose: Validates the generation of OIDC ID tokens, ensuring correct claims (iss, aud, sub, at_hash, nonce) and header (kid).
// Scope: Unit Test
// Security: OIDC compliance and cryptographic integrity of ID tokens
// Expected: A valid, signed JWT with all required OIDC claims.
func TestOIDC_Service_GenerateIDToken(t *testing.T) {
	issuer := "http://localhost:8080"
	s, err := NewService(issuer)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	userID := "user-123"
	tenantID := "tenant-456"
	clientID := "client-789"
	nonce := "random-nonce"
	accessToken := "raw-access-token"

	tokenString, err := s.GenerateIDToken(userID, tenantID, clientID, nonce, accessToken)
	if err != nil {
		t.Fatalf("failed to generate ID token: %v", err)
	}

	// Parse token to verify claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return &s.signingKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		t.Fatal("invalid token claims")
	}

	// Verify required claims
	if claims["iss"] != issuer {
		t.Errorf("expected iss %s, got %v", issuer, claims["iss"])
	}
	if claims["aud"] != clientID {
		t.Errorf("expected aud %s, got %v", clientID, claims["aud"])
	}
	if claims["nonce"] != nonce {
		t.Errorf("expected nonce %s, got %v", nonce, claims["nonce"])
	}

	// Verify at_hash
	if _, ok := claims["at_hash"]; !ok {
		t.Error("missing at_hash claim")
	}

	// Verify sub is not raw userID
	sub := claims["sub"].(string)
	if sub == userID {
		t.Error("sub claim should not be the raw userID")
	}

	// Verify kid in header
	if token.Header["kid"] != s.kid {
		t.Errorf("expected kid %s, got %v", s.kid, token.Header["kid"])
	}
}

// TestPurpose: Validates that the discovery metadata matches the service configuration and OIDC standards.
// Scope: Unit Test
// Security: Exposure of correct OIDC endpoints and supported algorithms
// Expected: A Metadata struct with correct issuer, JWKS URI and supported signing algorithms.
func TestOIDC_Service_GetDiscoveryMetadata(t *testing.T) {
	issuer := "https://auth.opentrusty.org"
	s, _ := NewService(issuer)

	meta := s.GetDiscoveryMetadata()

	if meta.Issuer != issuer {
		t.Errorf("expected issuer %s, got %s", issuer, meta.Issuer)
	}
	if !strings.Contains(meta.JWKSURI, "/jwks.json") {
		t.Errorf("invalid jwks_uri: %s", meta.JWKSURI)
	}
	if len(meta.IDTokenSigningAlgValuesSupported) == 0 || meta.IDTokenSigningAlgValuesSupported[0] != "RS256" {
		t.Errorf("RS256 should be supported")
	}
}

// TestPurpose: Validates that the JWKS (JSON Web Key Set) correctly exposes the public key for token verification.
// Scope: Unit Test
// Security: Public key distribution for signature verification
// Expected: A JWKS document containing the RSA public key components.
func TestOIDC_Service_GetJWKS(t *testing.T) {
	s, _ := NewService("http://localhost")

	jwks := s.GetJWKS()

	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(jwks.Keys))
	}

	key := jwks.Keys[0]
	if key.Kid != s.kid {
		t.Errorf("expected kid %s, got %s", s.kid, key.Kid)
	}
	if key.Kty != "RSA" {
		t.Errorf("expected kty RSA, got %s", key.Kty)
	}
	if key.Alg != "RS256" {
		t.Errorf("expected alg RS256, got %s", key.Alg)
	}
	if key.N == "" || key.E == "" {
		t.Error("RSA public key components (N, E) missing")
	}
}

package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service handles OpenID Connect specific logic (Phase II.2)
type Service struct {
	issuer     string
	signingKey *rsa.PrivateKey
	kid        string // Stable, deterministic Key ID
}

// DiscoveryMetadata represents OIDC Discovery metadata (OIDC Discovery Section 3)
type DiscoveryMetadata struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	JWKSURI                          string   `json:"jwks_uri"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                  []string `json:"scopes_supported"`
	GrantTypesSupported              []string `json:"grant_types_supported"`
}

// JWK represents a JSON Web Key (RFC 7517)
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set (RFC 7517)
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// NewService creates a new OIDC service
func NewService(issuer string) (*Service, error) {
	// For Phase II.1/II.2, we use a static generated key (no rotation/persistence yet)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// T6 (Phase II.2): Stable, deterministic kid
	// Generate kid using SHA-256 thumbprint of the N component (simplified)
	nBytes := key.PublicKey.N.Bytes()
	hash := sha256.Sum256(nBytes)
	kid := base64.RawURLEncoding.EncodeToString(hash[:16]) // First 16 bytes is enough for kid

	return &Service{
		issuer:     issuer,
		signingKey: key,
		kid:        kid,
	}, nil
}

// GetDiscoveryMetadata returns the OIDC configuration (OIDC Discovery Section 4)
func (s *Service) GetDiscoveryMetadata() DiscoveryMetadata {
	return DiscoveryMetadata{
		Issuer:                           s.issuer,
		AuthorizationEndpoint:            fmt.Sprintf("%s/oauth2/authorize", s.issuer),
		TokenEndpoint:                    fmt.Sprintf("%s/oauth2/token", s.issuer),
		JWKSURI:                          fmt.Sprintf("%s/jwks.json", s.issuer),
		ResponseTypesSupported:           []string{"code"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		ScopesSupported:                  []string{"openid"},
		GrantTypesSupported:              []string{"authorization_code", "refresh_token"},
	}
}

// GetJWKS returns the public keys in JWKS format (RFC 7517)
func (s *Service) GetJWKS() JWKS {
	pub := s.signingKey.PublicKey
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(bigIntToBytes(pub.E))

	return JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Use: "sig",
				Alg: "RS256",
				Kid: s.kid,
				N:   n,
				E:   e,
			},
		},
	}
}

// GenerateIDToken generates a signed id_token JWT (OIDC Core Section 2)
func (s *Service) GenerateIDToken(userID, tenantID, clientID, nonce, accessToken string) (string, error) {
	now := time.Now()

	subSource := fmt.Sprintf("%s:%s", tenantID, userID)
	hash := sha256.Sum256([]byte(subSource))
	sub := base64.RawURLEncoding.EncodeToString(hash[:])

	claims := jwt.MapClaims{
		"iss": s.issuer,
		"sub": sub,
		"aud": clientID,
		"exp": now.Add(5 * time.Minute).Unix(),
		"iat": now.Unix(),
	}

	// OIDC Core Section 3.1.2.1: Include nonce if provided
	if nonce != "" {
		claims["nonce"] = nonce
	}

	// OIDC Core Section 3.1.3.6: Compute at_hash if access_token is issued
	if accessToken != "" {
		// at_hash is base64url encoding of the left-most half of the hash
		// of the octets of the ASCII representation of the access_token.
		// For RS256, hash is SHA-256.
		atHash := sha256.Sum256([]byte(accessToken))
		leftHalf := atHash[:len(atHash)/2]
		claims["at_hash"] = base64.RawURLEncoding.EncodeToString(leftHalf)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Phase II.2: Include stable kid in header
	token.Header["kid"] = s.kid

	return token.SignedString(s.signingKey)
}

func bigIntToBytes(n int) []byte {
	if n == 0 {
		return []byte{0}
	}
	var res []byte
	for n > 0 {
		res = append([]byte{byte(n & 0xff)}, res...)
		n >>= 8
	}
	return res
}

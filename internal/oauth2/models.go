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
	"errors"
	"strings"
	"time"
)

// Domain errors (Internal)
var (
	ErrClientNotFound           = errors.New("client not found")
	ErrClientAlreadyExists      = errors.New("client already exists")
	ErrDomainInvalidRedirectURI = errors.New("invalid redirect URI")
	ErrDomainInvalidScope       = errors.New("invalid scope")
	ErrDomainInvalidGrantType   = errors.New("invalid grant type")
	ErrCodeExpired              = errors.New("authorization code expired")
	ErrCodeAlreadyUsed          = errors.New("authorization code already used")
	ErrCodeNotFound             = errors.New("authorization code not found")
	ErrDomainInvalidClient      = errors.New("invalid client credentials")
	ErrTokenExpired             = errors.New("token expired")
	ErrTokenRevoked             = errors.New("token revoked")
	ErrTokenNotFound            = errors.New("token not found")
)

const (
	ScopeOpenID = "openid"
	ScopeRoles  = "roles"
)

// Client represents an OAuth2 client application
type Client struct {
	ID                      string
	ClientID                string
	TenantID                string
	ClientSecretHash        string
	ClientName              string
	ClientURI               string
	LogoURI                 string
	RedirectURIs            []string
	AllowedScopes           []string
	GrantTypes              []string
	ResponseTypes           []string
	TokenEndpointAuthMethod string
	AccessTokenLifetime     int
	RefreshTokenLifetime    int
	IDTokenLifetime         int
	OwnerID                 string
	IsTrusted               bool
	IsActive                bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
	DeletedAt               *time.Time
}

// ValidateRedirectURI checks if the redirect URI is allowed for this client
func (c *Client) ValidateRedirectURI(redirectURI string) bool {
	for _, uri := range c.RedirectURIs {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

// ValidateScope checks if the requested scope is allowed for this client
func (c *Client) ValidateScope(requestedScope string) bool {
	if requestedScope == "" {
		return true
	}

	// Split space-separated scopes
	requestedScopes := strings.Fields(requestedScope)

	// Check if all requested scopes are allowed
	for _, reqScope := range requestedScopes {
		allowed := false
		for _, allowedScope := range c.AllowedScopes {
			if allowedScope == reqScope || allowedScope == "*" {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	return true
}

// AuthorizationCode represents a short-lived authorization code
type AuthorizationCode struct {
	ID                  string
	Code                string
	ClientID            string
	UserID              string
	RedirectURI         string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	UsedAt              *time.Time
	IsUsed              bool
	CreatedAt           time.Time
}

// IsExpired checks if the authorization code has expired
func (a *AuthorizationCode) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// AccessToken represents an OAuth2 access token
type AccessToken struct {
	ID        string
	TenantID  string
	TokenHash string
	ClientID  string
	UserID    string
	Scope     string
	TokenType string
	ExpiresAt time.Time
	RevokedAt *time.Time
	IsRevoked bool
	CreatedAt time.Time
}

// IsExpired checks if the access token has expired
func (a *AccessToken) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// RefreshToken represents an OAuth2 refresh token
type RefreshToken struct {
	ID            string
	TenantID      string
	TokenHash     string
	AccessTokenID string
	ClientID      string
	UserID        string
	Scope         string
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	IsRevoked     bool
	CreatedAt     time.Time
}

// IsExpired checks if the refresh token has expired
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// ClientRepository defines the interface for OAuth2 client persistence
type ClientRepository interface {
	// Create creates a new OAuth2 client
	Create(client *Client) error

	// GetByClientID retrieves a client by client_id
	GetByClientID(clientID string) (*Client, error)

	// GetByID retrieves a client by internal ID
	GetByID(id string) (*Client, error)

	// Update updates client information
	Update(client *Client) error

	// Delete soft-deletes a client
	Delete(id string) error

	// List retrieves all clients for an owner
	ListByOwner(ownerID string) ([]*Client, error)
}

// AuthorizationCodeRepository defines the interface for authorization code persistence
type AuthorizationCodeRepository interface {
	// Create creates a new authorization code
	Create(code *AuthorizationCode) error

	// GetByCode retrieves an authorization code
	GetByCode(code string) (*AuthorizationCode, error)

	// MarkAsUsed marks the code as used
	MarkAsUsed(code string) error

	// Delete deletes an authorization code
	Delete(code string) error

	// DeleteExpired deletes all expired authorization codes
	DeleteExpired() error
}

// AccessTokenRepository defines the interface for access token persistence
type AccessTokenRepository interface {
	// Create creates a new access token
	Create(token *AccessToken) error

	// GetByTokenHash retrieves an access token
	GetByTokenHash(tokenHash string) (*AccessToken, error)

	// Revoke revokes an access token
	Revoke(tokenHash string) error

	// DeleteExpired deletes all expired access tokens
	DeleteExpired() error
}

// RefreshTokenRepository defines the interface for refresh token persistence
type RefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(token *RefreshToken) error

	// GetByTokenHash retrieves a refresh token
	GetByTokenHash(tokenHash string) (*RefreshToken, error)

	// Revoke revokes a refresh token
	Revoke(tokenHash string) error

	// DeleteExpired deletes all expired refresh tokens
	DeleteExpired() error
}

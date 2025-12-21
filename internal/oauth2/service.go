package oauth2

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
)

// OIDCProvider defines the interface for OIDC integration (Phase II.3)
type OIDCProvider interface {
	GenerateIDToken(userID, tenantID, clientID, nonce, accessToken string) (string, error)
}

// Service provides OAuth2 business logic
type Service struct {
	clientRepo   ClientRepository
	codeRepo     AuthorizationCodeRepository
	accessRepo   AccessTokenRepository
	refreshRepo  RefreshTokenRepository
	auditLogger  audit.Logger
	oidcProvider OIDCProvider // Optional OIDC integration hook

	// Configuration
	authCodeLifetime     time.Duration
	accessTokenLifetime  time.Duration
	refreshTokenLifetime time.Duration
	encryptionKey        []byte // Master key for encrypting private keys in DB
}

// NewService creates a new OAuth2 service
func NewService(
	clientRepo ClientRepository,
	codeRepo AuthorizationCodeRepository,
	accessRepo AccessTokenRepository,
	refreshRepo RefreshTokenRepository,
	auditLogger audit.Logger,
	oidcProvider OIDCProvider,
) *Service {
	// Load encryption key from env or use default (dev only)
	encKey := []byte(os.Getenv("OPENID_KEY_ENCRYPTION_KEY"))
	if len(encKey) != 32 {
		encKey = make([]byte, 32)
		copy(encKey, []byte("insecure_dev_key_must_change_!!"))
	}

	return &Service{
		clientRepo:           clientRepo,
		codeRepo:             codeRepo,
		accessRepo:           accessRepo,
		refreshRepo:          refreshRepo,
		auditLogger:          auditLogger,
		oidcProvider:         oidcProvider,
		authCodeLifetime:     5 * time.Minute,
		accessTokenLifetime:  1 * time.Hour,
		refreshTokenLifetime: 30 * 24 * time.Hour,
		encryptionKey:        encKey,
	}
}

// AuthorizeRequest represents an OAuth2 authorization request
type AuthorizeRequest struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
}

// TokenRequest represents an OAuth2 token request
type TokenRequest struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	CodeVerifier string
	RefreshToken string
	Scope        string
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"` // Added for OIDC (Phase II.1)
	Scope        string `json:"scope,omitempty"`
}

// CreateClient registers a new OAuth2 client
func (s *Service) CreateClient(ctx context.Context, client *Client) error {
	if client.ID == "" {
		client.ID = generateID()
	}
	if client.ClientID == "" {
		client.ClientID = generateID()
	}

	if client.CreatedAt.IsZero() {
		client.CreatedAt = time.Now()
	}
	client.UpdatedAt = time.Now()

	return s.clientRepo.Create(client)
}

// ValidateAuthorizeRequest validates an authorization request (RFC 6749 Section 4.1.1)
func (s *Service) ValidateAuthorizeRequest(ctx context.Context, req *AuthorizeRequest) (*Client, error) {
	// 1. Validate Client (RFC 6749 Section 4.1.1)
	client, err := s.clientRepo.GetByClientID(req.ClientID)
	if err != nil {
		return nil, NewError(ErrInvalidRequest, "invalid client_id")
	}

	if !client.IsActive {
		return nil, NewError(ErrInvalidRequest, "client is disabled")
	}

	// 2. Validate Redirect URI (RFC 6749 Section 3.1.2)
	// Must be an exact match for registered URIs
	if !client.ValidateRedirectURI(req.RedirectURI) {
		return nil, NewError(ErrInvalidRequest, "invalid redirect_uri")
	}

	// 3. Validate Response Type (RFC 6749 Section 3.1.1)
	// Phase I.1 only supports 'code'
	if req.ResponseType != "code" {
		return nil, NewError(ErrUnsupportedGrantType, "response_type must be 'code'")
	}

	// 4. Validate Scope (RFC 6749 Section 3.3)
	if req.Scope != "" && !client.ValidateScope(req.Scope) {
		return nil, NewError(ErrInvalidScope, "invalid scope")
	}

	// 5. Validate PKCE Method (RFC 7636 Section 4.3)
	if req.CodeChallenge != "" {
		if req.CodeChallengeMethod != "" && req.CodeChallengeMethod != "plain" && req.CodeChallengeMethod != "S256" {
			return nil, NewError(ErrInvalidRequest, "transform algorithm not supported")
		}
	}

	return client, nil
}

// CreateAuthorizationCode creates a new authorization code (RFC 6749 Section 4.1.2)
func (s *Service) CreateAuthorizationCode(ctx context.Context, req *AuthorizeRequest, userID string) (*AuthorizationCode, error) {
	code := &AuthorizationCode{
		ID:                  generateID(),
		Code:                generateAuthorizationCode(),
		ClientID:            req.ClientID,
		UserID:              userID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		State:               req.State,
		Nonce:               req.Nonce,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		// Authorization codes MUST be short-lived (RFC 6749 Section 4.1.2 recommends < 10min)
		// We use 5 minutes per Phase I.1 requirements.
		ExpiresAt: time.Now().Add(5 * time.Minute),
		IsUsed:    false,
		CreatedAt: time.Now(),
	}

	if err := s.codeRepo.Create(code); err != nil {
		return nil, NewError(ErrServerError, "failed to persist authorization code")
	}

	return code, nil
}

func validatePKCE(challenge, method, verifier string) bool {
	// RFC 7636 Section 4.2: If the method is not specified, it defaults to "plain".
	if method == "" || method == "plain" {
		return challenge == verifier
	}

	// RFC 7636 Section 4.6
	if method == "S256" {
		hash := sha256.Sum256([]byte(verifier))
		computed := base64.RawURLEncoding.EncodeToString(hash[:])
		return challenge == computed
	}

	return false
}

// ExchangeCodeForToken exchanges an authorization code for tokens (RFC 6749 Section 4.1.3)
func (s *Service) ExchangeCodeForToken(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	// 1. Authenticate Client (RFC 6749 Section 3.2.1)
	client, err := s.ValidateClientCredentials(req.ClientID, req.ClientSecret)
	if err != nil {
		return nil, err
	}

	// 2. Validate Grant Type (RFC 6749 Section 4.1.3)
	if req.GrantType != "authorization_code" {
		return nil, NewError(ErrUnsupportedGrantType, "grant_type must be 'authorization_code'")
	}

	// 3. Retrieve and Validate Code (RFC 6749 Section 4.1.3)
	code, err := s.codeRepo.GetByCode(req.Code)
	if err != nil {
		return nil, NewError(ErrInvalidGrant, "authorization code not found")
	}

	if code.IsUsed {
		return nil, NewError(ErrInvalidGrant, "authorization code already used")
	}

	if code.IsExpired() {
		return nil, NewError(ErrInvalidGrant, "authorization code expired")
	}

	if code.ClientID != req.ClientID {
		return nil, NewError(ErrInvalidGrant, "client_id mismatch")
	}

	if code.RedirectURI != req.RedirectURI {
		return nil, NewError(ErrInvalidGrant, "redirect_uri mismatch")
	}

	// 4. PKCE Verification (RFC 7636 Section 4.6)
	if code.CodeChallenge != "" {
		if !validatePKCE(code.CodeChallenge, code.CodeChallengeMethod, req.CodeVerifier) {
			return nil, NewError(ErrInvalidGrant, "invalid code_verifier")
		}
	} else {
		// Public clients MUST use PKCE (if we enforce it per Rule 4, but for now we follow RFC)
		// RFC 7636 doesn't strictly MANDATE it for all, but highly recommends it.
		// If the code was issued without PKCE, we continue.
	}

	// 5. Mark code as used
	if err := s.codeRepo.MarkAsUsed(req.Code); err != nil {
		return nil, NewError(ErrServerError, "failed to invalidate authorization code")
	}

	// 6. Issue Access Token
	rawAccessToken := generateToken()
	accessToken := &AccessToken{
		ID:        generateID(),
		TokenHash: hashToken(rawAccessToken),
		ClientID:  client.ClientID,
		UserID:    code.UserID,
		Scope:     code.Scope,
		TokenType: "Bearer",
		ExpiresAt: time.Now().Add(time.Duration(client.AccessTokenLifetime) * time.Second),
		IsRevoked: false,
		CreatedAt: time.Now(),
	}

	if err := s.accessRepo.Create(accessToken); err != nil {
		return nil, NewError(ErrServerError, "failed to issue access token")
	}

	// 7. Issue Refresh Token (Optional, RFC 6749 Section 1.5)
	var refreshToken string
	allowedRefresh := false
	for _, gt := range client.GrantTypes {
		if gt == "refresh_token" {
			allowedRefresh = true
			break
		}
	}

	if allowedRefresh {
		rt := &RefreshToken{
			ID:            generateID(),
			TokenHash:     hashToken(generateToken()),
			AccessTokenID: accessToken.ID,
			ClientID:      client.ClientID,
			UserID:        code.UserID,
			Scope:         code.Scope,
			ExpiresAt:     time.Now().Add(time.Duration(client.RefreshTokenLifetime) * time.Second),
			IsRevoked:     false,
			CreatedAt:     time.Now(),
		}
		if err := s.refreshRepo.Create(rt); err == nil {
			refreshToken = rt.TokenHash
		}
	}

	// 8. Issue ID Token (OIDC Core Section 2)
	var idToken string
	if s.oidcProvider != nil && containsScope(code.Scope, "openid") {
		// Pass nonce and raw access token for at_hash computation (Phase II.3)
		it, err := s.oidcProvider.GenerateIDToken(code.UserID, client.TenantID, client.ClientID, code.Nonce, rawAccessToken)
		if err == nil {
			idToken = it
		} else {
			// Log error but don't fail the OAuth2 exchange if OIDC fails
			// This is a trade-off; some might prefer failing.
		}
	}

	// Audit token issuance
	s.auditLogger.Log(ctx, audit.Event{
		Type:     audit.TypeTokenIssued,
		TenantID: client.TenantID,
		ActorID:  code.UserID,
		Resource: "token",
		Metadata: map[string]any{
			"client_id": client.ClientID,
			"scope":     code.Scope,
			"has_rt":    refreshToken != "",
			"has_it":    idToken != "",
		},
	})

	return &TokenResponse{
		AccessToken:  rawAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    client.AccessTokenLifetime,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Scope:        code.Scope,
	}, nil
}

// RefreshAccessToken handles the refresh_token grant type (RFC 6749 Section 6)
func (s *Service) RefreshAccessToken(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	// 1. Authenticate Client (RFC 6749 Section 3.2.1)
	client, err := s.ValidateClientCredentials(req.ClientID, req.ClientSecret)
	if err != nil {
		return nil, err
	}

	// 2. Validate Refresh Token
	rt, err := s.refreshRepo.GetByTokenHash(req.RefreshToken)
	if err != nil {
		return nil, NewError(ErrInvalidGrant, "refresh token not found")
	}

	if rt.IsRevoked {
		return nil, NewError(ErrInvalidGrant, "refresh token revoked")
	}

	if rt.IsExpired() {
		return nil, NewError(ErrInvalidGrant, "refresh token expired")
	}

	if rt.ClientID != client.ClientID {
		return nil, NewError(ErrInvalidGrant, "client_id mismatch")
	}

	// 3. Issue New Access Token
	accessToken := &AccessToken{
		ID:        generateID(),
		TokenHash: hashToken(generateToken()),
		ClientID:  client.ClientID,
		UserID:    rt.UserID,
		Scope:     rt.Scope, // Scope SHOULD be same or subset (RFC 6749 Section 6)
		TokenType: "Bearer",
		ExpiresAt: time.Now().Add(time.Duration(client.AccessTokenLifetime) * time.Second),
		IsRevoked: false,
		CreatedAt: time.Now(),
	}

	if err := s.accessRepo.Create(accessToken); err != nil {
		return nil, NewError(ErrServerError, "failed to issue access token")
	}

	// Optional: Rotate refresh token (RFC 6749 Section 6)
	// For now, we keep the same refresh token to keep it simple as per Minimal Core.
	// But issuance of new RT is allowed.

	return &TokenResponse{
		AccessToken:  accessToken.TokenHash,
		TokenType:    "Bearer",
		ExpiresIn:    client.AccessTokenLifetime,
		RefreshToken: rt.TokenHash,
		Scope:        rt.Scope,
	}, nil
}

// ValidateClientCredentials validates client credentials (RFC 6749 Section 3.2.1)
func (s *Service) ValidateClientCredentials(clientID, clientSecret string) (*Client, error) {
	client, err := s.clientRepo.GetByClientID(clientID)
	if err != nil {
		return nil, NewError(ErrInvalidClient, "invalid client credentials")
	}

	if !client.IsActive {
		return nil, NewError(ErrInvalidClient, "client is disabled")
	}

	// Public Clients (RFC 6749 Section 2.1)
	if client.ClientSecretHash == "" {
		// Secret is empty for public clients
		return client, nil
	}

	// Confidential Clients - validate secret
	secretHash := hashClientSecret(clientSecret)
	if secretHash != client.ClientSecretHash {
		return nil, NewError(ErrInvalidClient, "invalid client credentials")
	}

	return client, nil
}

// ValidateAccessToken validates an access token
func (s *Service) ValidateAccessToken(ctx context.Context, tokenHash string) (*AccessToken, error) {
	token, err := s.accessRepo.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, ErrTokenNotFound
	}

	if token.IsRevoked {
		return nil, ErrTokenRevoked
	}

	if token.IsExpired() {
		return nil, ErrTokenExpired
	}

	return token, nil
}

// Helper functions

func (s *Service) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (s *Service) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// RevokeRefreshToken revokes a refresh token (Security Best Practice)
func (s *Service) RevokeRefreshToken(ctx context.Context, tokenHash string, clientID string) error {
	rt, err := s.refreshRepo.GetByTokenHash(tokenHash)
	if err != nil {
		return ErrTokenNotFound
	}

	if rt.ClientID != clientID {
		return NewError(ErrInvalidClient, "client_id mismatch")
	}

	return s.refreshRepo.Revoke(tokenHash)
}

func containsScope(scope, target string) bool {
	parts := strings.Split(scope, " ")
	for _, part := range parts {
		if part == target {
			return true
		}
	}
	return false
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateAuthorizationCode() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func hashClientSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GenerateClientSecret generates a new client secret
func GenerateClientSecret() string {
	return generateToken()
}

// HashClientSecret hashes a client secret for storage
func HashClientSecret(secret string) string {
	return hashClientSecret(secret)
}

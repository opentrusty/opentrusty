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
	"log/slog"
	"net/http"
	"strings"

	"github.com/opentrusty/opentrusty/internal/oauth2"
	"github.com/opentrusty/opentrusty/internal/oidc"
)

// Authorize endpoints
// @Summary OAuth2 Authorize Endpoint
// @Description Starts the authorization flow (RFC 6749)
// @Tags OAuth2
// @Accept json
// @Produce html
// @Param client_id query string true "Client ID"
// @Param redirect_uri query string true "Redirect URI"
// @Param response_type query string true "Response Type (must be 'code')"
// @Param scope query string false "Scopes"
// @Param state query string true "Random State"
// @Param nonce query string false "Nonce (OIDC)"
// @Param code_challenge query string false "PKCE Challenge"
// @Param code_challenge_method query string false "PKCE Method (S256)"
// @Success 302 {string} string "Redirects to callback or login"
// @Router /oauth2/authorize [get]
func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	req := &oauth2.AuthorizeRequest{
		ClientID:            query.Get("client_id"),
		RedirectURI:         query.Get("redirect_uri"),
		ResponseType:        query.Get("response_type"),
		Scope:               query.Get("scope"),
		State:               query.Get("state"),
		Nonce:               query.Get("nonce"),
		CodeChallenge:       query.Get("code_challenge"),
		CodeChallengeMethod: query.Get("code_challenge_method"),
	}

	// Validate request parameters first
	_, err := h.oauth2Service.ValidateAuthorizeRequest(r.Context(), req)
	if err != nil {
		slog.ErrorContext(r.Context(), "invalid authorize request",
			"error", err,
			"client_id", req.ClientID,
			"redirect_uri", req.RedirectURI,
		)

		// If redirect URI is valid, redirect back with error
		// Otherwise, show error page (for now, JSON error)
		if oe, ok := err.(*oauth2.Error); ok {
			// If redirect_uri is invalid, we can't redirect
			if oe.Code == oauth2.ErrInvalidRequest && strings.Contains(oe.Description, "redirect") {
				h.respondOAuthError(w, err)
				return
			}

			// Redirect back with error
			redirectURL := addQueryParams(req.RedirectURI, map[string]string{
				"error":             oe.Code,
				"error_description": oe.Description,
				"state":             req.State,
			})
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}

		h.respondOAuthError(w, err)
		return
	}

	// Check if user is authenticated
	userID := GetUserID(r.Context())
	if userID == "" {
		// Store authorization request in session or cookie (simplified for now)
		// Redirect to login page with return_to param
		// For API-first approach, we return 401 and expect client to handle login
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// TODO: Display consent page if needed
	// For now, auto-approve if user is authenticated

	// Generate authorization code
	code, err := h.oauth2Service.CreateAuthorizationCode(r.Context(), req, userID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to create authorization code", "error", err)
		redirectURL := addQueryParams(req.RedirectURI, map[string]string{
			"error": "server_error",
			"state": req.State,
		})
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// Redirect back with code
	redirectURL := addQueryParams(req.RedirectURI, map[string]string{
		"code":  code.Code,
		"state": req.State,
	})
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// Token endpoint
// @Summary OAuth2 Token Endpoint
// @Description Exchange code for access token (RFC 6749)
// @Tags OAuth2
// @Accept x-www-form-urlencoded
// @Produce json
// @Param grant_type formData string true "Grant Type (authorization_code or refresh_token)"
// @Param code formData string false "Authorization Code (for authorization_code grant)"
// @Param redirect_uri formData string false "Redirect URI"
// @Param client_id formData string false "Client ID (if not Basic Auth)"
// @Param client_secret formData string false "Client Secret (if not Basic Auth)"
// @Param code_verifier formData string false "PKCE Verifier"
// @Param refresh_token formData string false "Refresh Token (for refresh_token grant)"
// @Param scope formData string false "Scope"
// @Success 200 {object} oauth2.TokenResponse
// @Failure 400 {object} oauth2.Error
// @Failure 401 {object} oauth2.Error
// @Router /oauth2/token [post]
func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Extract credentials
	clientID := r.Form.Get("client_id")
	clientSecret := r.Form.Get("client_secret")

	// Support Basic Auth (RFC 6749 Section 2.3.1)
	if clientID == "" {
		username, password, ok := r.BasicAuth()
		if ok {
			clientID = username
			clientSecret = password
		}
	}

	req := &oauth2.TokenRequest{
		GrantType:    r.Form.Get("grant_type"),
		Code:         r.Form.Get("code"),
		RedirectURI:  r.Form.Get("redirect_uri"),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CodeVerifier: r.Form.Get("code_verifier"), // RFC 7636 Section 4.5
		RefreshToken: r.Form.Get("refresh_token"), // RFC 6749 Section 6
		Scope:        r.Form.Get("scope"),
	}

	var resp *oauth2.TokenResponse
	var err error

	switch req.GrantType {
	case "authorization_code":
		// RFC 6749 Section 4.1.3
		resp, err = h.oauth2Service.ExchangeCodeForToken(r.Context(), req)
	case "refresh_token":
		// RFC 6749 Section 6
		resp, err = h.oauth2Service.RefreshAccessToken(r.Context(), req)
	default:
		h.respondOAuthError(w, oauth2.NewError(oauth2.ErrUnsupportedGrantType, "unsupported grant_type"))
		return
	}

	if err != nil {
		slog.ErrorContext(r.Context(), "token request failed", "error", err, "grant_type", req.GrantType)
		h.respondOAuthError(w, err)
		return
	}

	// Prevent caching (RFC 6749 Section 5.1)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	respondJSON(w, http.StatusOK, resp)
}

// Revoke handle the token revocation request (RFC 7009)
// @Summary Revoke Token
// @Description Revoke a refresh token (RFC 7009)
// @Tags OAuth2
// @Accept x-www-form-urlencoded
// @Produce json
// @Param token formData string true "Token to revoke"
// @Param client_id formData string false "Client ID"
// @Param client_secret formData string false "Client Secret"
// @Success 200 {string} string "OK"
// @Failure 400 {object} oauth2.Error
// @Router /oauth2/revoke [post]
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.respondOAuthError(w, oauth2.NewError(oauth2.ErrInvalidRequest, "invalid request"))
		return
	}

	clientID := r.Form.Get("client_id")
	clientSecret := r.Form.Get("client_secret")

	// Support Basic Auth
	if clientID == "" {
		username, password, ok := r.BasicAuth()
		if ok {
			clientID = username
			clientSecret = password
		}
	}

	token := r.Form.Get("token")
	if token == "" {
		h.respondOAuthError(w, oauth2.NewError(oauth2.ErrInvalidRequest, "missing token"))
		return
	}

	// Validate client first
	_, err := h.oauth2Service.ValidateClientCredentials(clientID, clientSecret)
	if err != nil {
		h.respondOAuthError(w, err)
		return
	}

	// Revoke (we only support refresh tokens for now in Phase I.2, but should handle access tokens too if hash is stored)
	// For now, we attempt to revoke it as a refresh token
	_ = h.oauth2Service.RevokeRefreshToken(r.Context(), token, clientID)

	// RFC 7009 Section 2.2: The authorization server responds with an HTTP 200 OK
	// regardless of whether the token was already revoked or the token was invalid.
	w.WriteHeader(http.StatusOK)
}

// Helper to add query params to URL
func addQueryParams(rawURL string, params map[string]string) string {
	if strings.Contains(rawURL, "?") {
		rawURL += "&"
	} else {
		rawURL += "?"
	}

	var parts []string
	for k, v := range params {
		parts = append(parts, k+"="+v)
	}

	return rawURL + strings.Join(parts, "&")
}

// respondOAuthError serializes a protocol error into HTTP response.
// Fix for B-PROTOCOL-01 (Rule 2 - Domain-Driven Translation)
func (h *Handler) respondOAuthError(w http.ResponseWriter, err error) {
	if oauthErr, ok := err.(*oauth2.Error); ok {
		status := http.StatusBadRequest
		if oauthErr.Code == oauth2.ErrInvalidClient {
			status = http.StatusUnauthorized
		}
		if oauthErr.Code == oauth2.ErrServerError {
			status = http.StatusInternalServerError
		}
		respondJSON(w, status, oauthErr)
		return
	}

	if oidcErr, ok := err.(*oidc.Error); ok {
		respondJSON(w, http.StatusBadRequest, oidcErr)
		return
	}

	// Fallback for internal errors (opaque)
	respondJSON(w, http.StatusInternalServerError, oauth2.NewError(oauth2.ErrServerError, "internal server error"))
}

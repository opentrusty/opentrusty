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
	"encoding/json"
	"net/http"

	"github.com/opentrusty/opentrusty/internal/oauth2"
)

// RegisterClientRequest represents the data for registering a new OAuth2 client
type RegisterClientRequest struct {
	ClientName              string   `json:"client_name" binding:"required" example:"My Application"`
	RedirectURIs            []string `json:"redirect_uris" binding:"required" example:"[\"http://localhost:3000/callback\"]"`
	AllowedScopes           []string `json:"allowed_scopes" example:"[\"openid\", \"profile\"]"`
	GrantTypes              []string `json:"grant_types" example:"[\"authorization_code\", \"refresh_token\"]"`
	ResponseTypes           []string `json:"response_types" example:"[\"code\"]"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method" example:"client_secret_basic"`
}

// RegisterClientResponse represents the response after registering a client
type RegisterClientResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	ClientName   string `json:"client_name"`
}

// RegisterClient handles OAuth2 client registration
// @Summary Register Client
// @Description Register a new OAuth2 client for the tenant
// @Tags OAuth2
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body RegisterClientRequest true "Client Data"
// @Success 201 {object} RegisterClientResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants/{tenantID}/oauth2/clients [post]
func (h *Handler) RegisterClient(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	var req RegisterClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Generate client secret for confidential clients
	clientSecret := ""
	clientSecretHash := ""
	if req.TokenEndpointAuthMethod != "none" {
		clientSecret = oauth2.GenerateClientSecret()
		clientSecretHash = oauth2.HashClientSecret(clientSecret)
	}

	client := &oauth2.Client{
		TenantID:                tenantID,
		ClientName:              req.ClientName,
		ClientSecretHash:        clientSecretHash,
		RedirectURIs:            req.RedirectURIs,
		AllowedScopes:           req.AllowedScopes,
		GrantTypes:              req.GrantTypes,
		ResponseTypes:           req.ResponseTypes,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    2592000,
		IDTokenLifetime:         3600,
		IsActive:                true,
	}

	if len(client.AllowedScopes) == 0 {
		client.AllowedScopes = []string{"openid"}
	}
	if len(client.GrantTypes) == 0 {
		client.GrantTypes = []string{"authorization_code"}
	}
	if len(client.ResponseTypes) == 0 {
		client.ResponseTypes = []string{"code"}
	}

	if err := h.oauth2Service.CreateClient(r.Context(), client); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to register client: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, RegisterClientResponse{
		ClientID:     client.ClientID,
		ClientSecret: clientSecret,
		ClientName:   client.ClientName,
	})
}

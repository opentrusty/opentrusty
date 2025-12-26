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

	"github.com/go-chi/chi/v5"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
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
	tenantID := GetTenantID(r.Context())

	// Authorization Check: Tenant Admin required to register clients
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantManageClients)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "client management access required")
		return
	}

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

// ListClients handles listing OAuth2 clients for a tenant
func (h *Handler) ListClients(w http.ResponseWriter, r *http.Request) {
	tenantID := GetTenantID(r.Context())

	clients, err := h.oauth2Service.ListClients(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list clients: "+err.Error())
		return
	}

	// Permission check is already done by route wrapper in handlers.go if we moved it there,
	// but currently handlers.go routes are under AuthMiddleware and TenantMiddleware.
	// Actually, I should check permission here to be safe and specific.
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantManageClients)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "client management access required")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"clients": clients,
		"total":   len(clients),
	})
}

// GetClient handles retrieving a specific OAuth2 client
func (h *Handler) GetClient(w http.ResponseWriter, r *http.Request) {
	tenantID := GetTenantID(r.Context())
	clientID := chi.URLParam(r, "clientID")

	client, err := h.oauth2Service.GetClient(r.Context(), clientID)
	if err != nil {
		respondError(w, http.StatusNotFound, "client not found")
		return
	}

	if client.TenantID != tenantID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	respondJSON(w, http.StatusOK, client)
}

// DeleteClient handles deleting an OAuth2 client
func (h *Handler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	tenantID := GetTenantID(r.Context())
	clientID := chi.URLParam(r, "clientID")

	// Authorization check
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantManageClients)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "client management access required")
		return
	}

	client, err := h.oauth2Service.GetClient(r.Context(), clientID)
	if err != nil {
		respondError(w, http.StatusNotFound, "client not found")
		return
	}

	if client.TenantID != tenantID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := h.oauth2Service.DeleteClient(r.Context(), clientID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete client")
		return
	}

	if h.auditLogger != nil {
		h.auditLogger.Log(r.Context(), audit.Event{
			Type:     "client_deleted",
			TenantID: tenantID,
			ActorID:  GetUserID(r.Context()),
			Resource: "oauth2_client",
			Metadata: map[string]any{"client_id": clientID},
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegenerateClientSecret handles regenerating a client secret
func (h *Handler) RegenerateClientSecret(w http.ResponseWriter, r *http.Request) {
	tenantID := GetTenantID(r.Context())
	clientID := chi.URLParam(r, "clientID")

	client, err := h.oauth2Service.GetClient(r.Context(), clientID)
	if err != nil {
		respondError(w, http.StatusNotFound, "client not found")
		return
	}

	if client.TenantID != tenantID {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	if client.TokenEndpointAuthMethod == "none" {
		respondError(w, http.StatusBadRequest, "cannot regenerate secret for public client")
		return
	}

	newSecret := oauth2.GenerateClientSecret()
	client.ClientSecretHash = oauth2.HashClientSecret(newSecret)

	if err := h.oauth2Service.UpdateClient(r.Context(), client); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update client secret")
		return
	}

	h.auditLogger.Log(r.Context(), audit.Event{
		Type:     "client_secret_regenerated",
		TenantID: tenantID,
		ActorID:  GetUserID(r.Context()),
		Resource: "oauth2_client",
		Metadata: map[string]any{"client_id": clientID},
	})

	respondJSON(w, http.StatusOK, map[string]string{
		"client_secret": newSecret,
	})
}

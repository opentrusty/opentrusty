package http

import (
	"net/http"

	"github.com/opentrusty/opentrusty/internal/oidc"
)

// Discovery returns the OpenID Connect metadata (OIDC Discovery Section 4)
// @Summary OIDC Discovery
// @Description Returns OpenID Connect configuration metadata
// @Tags OIDC
// @Produce json
// @Success 200 {object} oidc.DiscoveryMetadata
// @Router /.well-known/openid-configuration [get]
func (h *Handler) Discovery(w http.ResponseWriter, r *http.Request) {
	// For Phase II.2, we assume h.oauth2Service has an oidcProvider that implements a new interface
	// or we add oidcService directly to the handler if needed.
	// But according to our architecture, h.oauth2Service.OIDCHook could be used.
	// However, for clean separation, we'll cast it or handle it appropriately.

	// Better yet, we can add oidcService to the Handler struct.
	// Since OIDC logic should stay in internal/oidc, we'll use a specific service field.

	var metadata oidc.DiscoveryMetadata = h.oidcService.GetDiscoveryMetadata()

	// OIDC Discovery Section 4.2: Content-Type MUST be application/json
	w.Header().Set("Content-Type", "application/json")
	respondJSON(w, http.StatusOK, metadata)
}

// JWKS returns the JSON Web Key Set (RFC 7517)
// @Summary JWKS
// @Description Returns the JSON Web Key Set for verify signing
// @Tags OIDC
// @Produce json
// @Success 200 {object} oidc.JWKS
// @Router /jwks.json [get]
func (h *Handler) JWKS(w http.ResponseWriter, r *http.Request) {
	var jwks oidc.JWKS = h.oidcService.GetJWKS()

	// RFC 7517 Section 8.1: Content-Type SHOULD be application/jwk-set+json
	// but OIDC clients often expect application/json.
	w.Header().Set("Content-Type", "application/json")
	respondJSON(w, http.StatusOK, jwks)
}

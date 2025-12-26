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
	"net/http"

	"github.com/opentrusty/opentrusty/internal/authz"
)

// ListTenants handles listing all tenants
// @Summary List Tenants
// @Description List all platform tenants (Platform Admin Only)
// @Tags Tenant
// @Produce json
// @Security CookieAuth
// @Success 200 {array} tenant.Tenant
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants [get]
func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	// 1. Authorization Check: Platform Admin required
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "platform admin administrative access required")
		return
	}

	tenants, err := h.tenantService.ListTenants(r.Context(), 100, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list tenants")
		return
	}

	respondJSON(w, http.StatusOK, tenants)
}

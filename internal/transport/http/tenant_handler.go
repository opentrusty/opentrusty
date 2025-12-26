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
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/identity"
	"github.com/opentrusty/opentrusty/internal/tenant"
)

// CreateTenantRequest represents tenant creation data
type CreateTenantRequest struct {
	Name string `json:"name" binding:"required" example:"My Corporation"`
}

// CreateTenant handles tenant creation
// @Summary Create Tenant
// @Description Create a new platform tenant (Platform Admin Only)
// @Tags Tenant
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body CreateTenantRequest true "Tenant Data"
// @Success 201 {object} tenant.Tenant
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants [post]
func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	// 1. Authorization Check: Platform Admin required
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "platform admin administrative access required")
		return
	}

	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.tenantService.CreateTenant(r.Context(), req.Name, userID)
	if err != nil {
		// Map domain errors to HTTP status codes
		if err == tenant.ErrInvalidTenantName {
			respondError(w, http.StatusBadRequest, "invalid tenant name")
			return
		}
		if err == tenant.ErrTenantAlreadyExists {
			respondError(w, http.StatusConflict, "tenant with this name already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create tenant")
		return
	}

	// 3. Security Event: Minimal audit log for tenant creation
	h.auditLogger.Log(r.Context(), audit.Event{
		Type:     audit.TypeTenantCreated,
		ActorID:  userID,
		Resource: audit.ResourceTenant,
		Metadata: map[string]any{
			audit.AttrTenantID:   t.ID,
			audit.AttrTenantName: t.Name,
		},
	})

	respondJSON(w, http.StatusCreated, t)
}

// ProvisionUserRequest represents user provisioning data
type ProvisionUserRequest struct {
	Email      string `json:"email" binding:"required" example:"user@example.com"`
	Password   string `json:"password" example:"secret123"`
	GivenName  string `json:"given_name" example:"John"`
	FamilyName string `json:"family_name" example:"Doe"`
	Role       string `json:"role" example:"admin"`
}

// ProvisionTenantUser handles provisioning a user in a tenant (Create + Assign Role)
// @Summary Provision Tenant User
// @Description Create a user and assign a role within a tenant
// @Tags Tenant
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body ProvisionUserRequest true "User Data"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants/{tenantID}/users [post]
func (h *Handler) ProvisionTenantUser(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		respondError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	var req ProvisionUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 1. Authorization Check: Tenant Admin or Platform Admin required
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantManageUsers)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant administrative access required")
		return
	}

	if req.Role == "" {
		req.Role = tenant.RoleTenantMember
	}

	// 1. Check if user exists
	user, err := h.identityService.GetByEmail(r.Context(), tenantID, req.Email)
	if err == nil && user != nil {
		// User exists, just assign role
	} else if err == identity.ErrUserNotFound {
		// Create user
		if req.Password == "" {
			respondError(w, http.StatusBadRequest, "password is required for new user")
			return
		}
		profile := identity.Profile{
			GivenName:  req.GivenName,
			FamilyName: req.FamilyName,
			FullName:   req.GivenName + " " + req.FamilyName,
		}
		user, err = h.identityService.ProvisionIdentity(r.Context(), tenantID, req.Email, profile)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to provision user: "+err.Error())
			return
		}

		if err := h.identityService.AddPassword(r.Context(), user.ID, req.Password); err != nil {
			respondError(w, http.StatusBadRequest, "failed to set password: "+err.Error())
			return
		}
	} else {
		slog.ErrorContext(r.Context(), "failed to check user", "error", err, "tenant_id", tenantID, "email", req.Email)
		respondError(w, http.StatusInternalServerError, "failed to check user: "+err.Error())
		return
	}

	slog.DebugContext(r.Context(), "user checked/provisioned", "user_id", user.ID, "tenant_id", tenantID)

	// 2. Assign role
	// Identify who is granting the role (current user)
	granterID := GetUserID(r.Context())

	err = h.tenantService.AssignRole(r.Context(), tenantID, user.ID, req.Role, granterID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to assign role", "error", err, "tenant_id", tenantID, "user_id", user.ID, "role", req.Role)
		if err == tenant.ErrRoleAlreadyExists {
			respondError(w, http.StatusConflict, "role already assigned")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to assign role: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		JSONKeyUserID: user.ID,
		JSONKeyRole:   req.Role,
	})
}

// AssignRoleRequest represents role assignment data
type AssignRoleRequest struct {
	Role string `json:"role" binding:"required" example:"member"`
}

// AssignTenantRole handles assigning a role to an existing user
// @Summary Assign Role
// @Description Assign a role to a user within a tenant
// @Tags Tenant
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param tenantID path string true "Tenant ID"
// @Param userID path string true "User ID"
// @Param request body AssignRoleRequest true "Role Data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants/{tenantID}/users/{userID}/roles [post]
func (h *Handler) AssignTenantRole(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	userID := chi.URLParam(r, "userID")

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Authorization Check: Tenant Admin or Platform Admin required
	granterID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), granterID, authz.ScopeTenant, &tenantID, authz.PermTenantManageUsers)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant administrative access required")
		return
	}

	err = h.tenantService.AssignRole(r.Context(), tenantID, userID, req.Role, granterID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to assign role")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

// RevokeTenantRole handles revoking a role
// @Summary Revoke Role
// @Description Revoke a role from a user within a tenant
// @Tags Tenant
// @Produce json
// @Security CookieAuth
// @Param tenantID path string true "Tenant ID"
// @Param userID path string true "User ID"
// @Param role path string true "Role"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants/{tenantID}/users/{userID}/roles/{role} [delete]
func (h *Handler) RevokeTenantRole(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	userID := chi.URLParam(r, "userID")
	role := chi.URLParam(r, "role")

	// 1. Authorization Check: Tenant Admin or Platform Admin required
	actorID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), actorID, authz.ScopeTenant, &tenantID, authz.PermTenantManageUsers)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant administrative access required")
		return
	}

	err = h.tenantService.RevokeRole(r.Context(), tenantID, userID, role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// ListTenantUsers lists users with roles
// @Summary List Tenant Users
// @Description List all users and their roles in a tenant
// @Tags Tenant
// @Produce json
// @Security CookieAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {array} tenant.TenantUserRole
// @Failure 500 {object} map[string]string
// @Router /tenants/{tenantID}/users [get]
func (h *Handler) ListTenantUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")

	// 1. Authorization Check: Tenant View permission required
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantView)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant view access required")
		return
	}

	roles, err := h.tenantService.GetTenantUsers(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, roles)
}

// AssignOwnerRequest represents tenant owner assignment data
type AssignOwnerRequest struct {
	UserID string `json:"user_id" binding:"required" example:"uuid"`
}

// AssignTenantOwner handles assigning a primary owner (tenant_owner role) to a tenant
// @Summary Assign Tenant Owner
// @Description Assign the 'tenant_owner' role to a user (Platform Admin Only)
// @Tags Tenant
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body AssignOwnerRequest true "Owner Data"
// @Success 200 {object} map[string]string
// @Router /tenants/{tenantID}/owners [post]
func (h *Handler) AssignTenantOwner(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")

	// 1. Authorization: Platform Admin only
	actorID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), actorID, authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "platform administrative access required")
		return
	}

	var req AssignOwnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 2. Assign 'tenant_owner' role
	err = h.tenantService.AssignRole(r.Context(), tenantID, req.UserID, tenant.RoleTenantOwner, actorID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to assign tenant owner: "+err.Error())
		return
	}

	// 3. Audit Log
	h.auditLogger.Log(r.Context(), audit.Event{
		Type:     audit.TypeRoleAssigned,
		TenantID: tenantID,
		ActorID:  actorID,
		Resource: audit.ResourceTenant,
		Metadata: map[string]any{
			audit.AttrRoleID: tenant.RoleTenantOwner,
			"target_user_id": req.UserID,
		},
	})

	respondJSON(w, http.StatusOK, map[string]string{"status": "owner_assigned"})
}

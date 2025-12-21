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
	"github.com/opentrusty/opentrusty/internal/identity"
	"github.com/opentrusty/opentrusty/internal/tenant"
)

// CreateTenantRequest represents tenant creation data
type CreateTenantRequest struct {
	ID   string `json:"id" binding:"required" example:"tenant-1"`
	Name string `json:"name" binding:"required" example:"My Corporation"`
}

// CreateTenant handles tenant creation
// @Summary Create Tenant
// @Description Create a new tenant
// @Tags Tenant
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body CreateTenantRequest true "Tenant Data"
// @Success 201 {object} tenant.Tenant
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants [post]
func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.tenantService.CreateTenant(r.Context(), req.ID, req.Name)
	if err != nil {
		// TODO: specific error handling
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

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
	granterID := ""
	if uid, ok := r.Context().Value("user_id").(string); ok {
		granterID = uid
	}

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
		"user_id": user.ID,
		"role":    req.Role,
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

	granterID := ""
	if uid, ok := r.Context().Value("user_id").(string); ok {
		granterID = uid
	}

	err := h.tenantService.AssignRole(r.Context(), tenantID, userID, req.Role, granterID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
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

	err := h.tenantService.RevokeRole(r.Context(), tenantID, userID, role)
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

	roles, err := h.tenantService.GetTenantUsers(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, roles)
}

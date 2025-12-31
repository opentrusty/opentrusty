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
	"crypto/rand"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/identity"
	"github.com/opentrusty/opentrusty/internal/tenant"
)

// CreateTenantRequest represents tenant creation data
type CreateTenantRequest struct {
	Name       string `json:"name" binding:"required" example:"My Corporation"`
	AdminEmail string `json:"admin_email,omitempty" example:"admin@example.com"`
	AdminName  string `json:"admin_name,omitempty" example:"Admin User"`
}

// CreateTenant handles tenant creation
// @Summary Create Tenant
// @Description Create a new platform tenant (Platform Admin Only). Optionally provision an initial tenant admin.
// @Tags Tenant
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body CreateTenantRequest true "Tenant Data"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants [post]
func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	// 1. Authorization Check: Platform Admin required
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "platform administrative access required")
		return
	}

	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 2. Generate Bootstrap Password if not provided
	adminPassword := "" // If provided in request? Add to CreateTenantRequest
	// For now, we'll generate one if not provided to ensure Requirement 3
	adminPassword, _ = generateRandomPassword(16)

	t, err := h.tenantService.CreateTenant(r.Context(), req.Name, req.AdminEmail, adminPassword, userID)
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
		respondError(w, http.StatusInternalServerError, "failed to create tenant: "+err.Error())
		return
	}

	response := map[string]any{
		"id":               t.ID,
		"name":             t.Name,
		"status":           t.Status,
		"created_at":       t.CreatedAt,
		"updated_at":       t.UpdatedAt,
		"admin_email":      req.AdminEmail,
		"admin_password":   adminPassword,
		"password_warning": "This password will not be shown again. Please copy it now.",
	}

	respondJSON(w, http.StatusCreated, response)
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

	// 1. Authorization Check: Tenant Admin ONLY (Platform Admin excluded)
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantManageUsers)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant administrative access required")
		return
	}

	if req.Role == "" {
		req.Role = tenant.RoleTenantMember
	}

	// 1. Check if user exists globally
	user, err := h.identityService.GetByEmail(r.Context(), req.Email)
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
		user, err = h.identityService.ProvisionIdentity(r.Context(), req.Email, profile)
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

	// 3. Audit Log for User Provisioning
	h.auditLogger.Log(r.Context(), audit.Event{
		Type:       audit.TypeUserCreated,
		TenantID:   tenantID,
		ActorID:    granterID,
		Resource:   audit.ResourceUser,
		TargetID:   user.ID,
		TargetName: req.Email,
		Metadata: map[string]any{
			audit.AttrEmail:  req.Email,
			audit.AttrRoleID: req.Role,
		},
	})

	// 4. Return the provisioned user info
	response := map[string]any{
		JSONKeyUserID: user.ID,
		JSONKeyRole:   req.Role,
	}
	// Only include password if we just created the user with one
	if req.Password != "" {
		response["password"] = req.Password
		response["password_warning"] = "This password will not be shown again. Please copy it now."
	}

	respondJSON(w, http.StatusOK, response)
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

	// Authorization Check: Tenant Admin ONLY (Platform Admin excluded)
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

	// 1. Authorization Check: Tenant Admin ONLY (Platform Admin excluded)
	actorID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), actorID, authz.ScopeTenant, &tenantID, authz.PermTenantManageUsers)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant administrative access required")
		return
	}

	err = h.tenantService.RevokeRole(r.Context(), tenantID, userID, role, actorID)
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

	// 1. Authorization Check: Tenant View permission required (Must be a member of the tenant)
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

// UpdateTenantUserRequest represents user update data
type UpdateTenantUserRequest struct {
	Nickname string `json:"nickname"`
}

// UpdateTenantUser handles updating a user's profile in a tenant
func (h *Handler) UpdateTenantUser(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	userID := chi.URLParam(r, "userID")

	// 1. Authorization: Tenant Admin ONLY
	actorID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), actorID, authz.ScopeTenant, &tenantID, authz.PermTenantManageUsers)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant administrative access required")
		return
	}

	var req UpdateTenantUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Fetch current profile to avoid overwriting other fields (like Picture)
	user, err := h.identityService.GetUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	profile := user.Profile
	profile.Nickname = req.Nickname

	if err := h.tenantService.UpdateUser(r.Context(), tenantID, userID, profile, actorID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update user: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
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

func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

// GetTenantMetrics returns summary statistics for a tenant
func (h *Handler) GetTenantMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		respondError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	// 1. Authorization: Tenant View permission (Tenant Owner/Admin/Member)
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantView)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant view access required")
		return
	}

	// 2. Fetch Users Count - Includes owners, admins, and members
	users, err := h.tenantService.GetTenantUsers(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	// 3. Fetch Clients Count
	clients, err := h.oauth2Service.ListClients(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch clients")
		return
	}

	// 4. Fetch Audit Count (24h)
	yesterday := time.Now().Add(-24 * time.Hour)
	_, auditCount, err := h.auditRepo.List(r.Context(), audit.Filter{
		TenantID:  &tenantID,
		StartDate: &yesterday,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch audit count", "error", err, "tenantID", tenantID)
		// We don't fail the whole request for audit count
	}

	respondJSON(w, http.StatusOK, tenant.TenantMetrics{
		TotalUsers:    len(users),
		TotalClients:  len(clients),
		AuditCount24h: auditCount,
	})
}

// GetTenant returns details of a specific tenant
// @Summary Get tenant details
// @Description Get details of a specific tenant (requires tenant view permission)
// @Tags Tenant
// @Produce json
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} map[string]any
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /tenants/{tenantID} [get]
// @Security SessionCookie
func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		respondError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantView)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "tenant view access required")
		return
	}

	t, err := h.tenantService.GetTenant(r.Context(), tenantID)
	if err != nil {
		slog.ErrorContext(r.Context(), "GetTenant failed", "error", err, "tenantID", tenantID)
		respondError(w, http.StatusNotFound, "tenant not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":         t.ID,
		"name":       t.Name,
		"status":     t.Status,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	})
}

// UpdateTenantRequest represents tenant update data
type UpdateTenantRequest struct {
	Name string `json:"name"`
}

// UpdateTenant handles tenant updates
func (h *Handler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		respondError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	// Authorization: Platform Admin Only
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "platform administrative access required")
		return
	}

	var req UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedTenant, err := h.tenantService.UpdateTenant(r.Context(), tenantID, req.Name, userID)
	if err != nil {
		if err == tenant.ErrTenantNotFound {
			respondError(w, http.StatusNotFound, "tenant not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update tenant")
		return
	}

	respondJSON(w, http.StatusOK, updatedTenant)
}

// DeleteTenant handles tenant deletion
func (h *Handler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		respondError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	// Authorization: Platform Admin Only
	userID := GetUserID(r.Context())
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "platform administrative access required")
		return
	}

	err = h.tenantService.DeleteTenant(r.Context(), tenantID, userID)
	if err != nil {
		if err == tenant.ErrTenantNotFound {
			respondError(w, http.StatusNotFound, "tenant not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete tenant")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

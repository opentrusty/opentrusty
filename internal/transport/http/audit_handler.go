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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
)

// AuditEventResponse represents an audit event in API responses
type AuditEventResponse struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	TenantID  string         `json:"tenant_id,omitempty"`
	ActorID   string         `json:"actor_id"`
	Resource  string         `json:"resource"`
	IPAddress string         `json:"ip_address,omitempty"`
	UserAgent string         `json:"user_agent,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt string         `json:"created_at"`
}

// ListAuditEventsResponse represents the audit events list response
type ListAuditEventsResponse struct {
	Events []AuditEventResponse `json:"events"`
	Total  int                  `json:"total"`
}

// ListTenantAuditEvents lists audit events for a specific tenant
// @Summary List tenant audit events
// @Description List audit events for a specific tenant (requires tenant admin or platform admin)
// @Tags Audit
// @Produce json
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} ListAuditEventsResponse
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants/{tenantID}/audit [get]
// @Security SessionCookie
func (h *Handler) ListTenantAuditEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		respondError(w, http.StatusBadRequest, "tenant ID is required")
		return
	}

	userID := GetUserID(r.Context())

	// Authorization: Check if user can view audit logs for this tenant
	// Either tenant admin for this tenant, or platform admin
	isTenantAdmin, _ := h.authzService.HasPermission(r.Context(), userID, authz.ScopeTenant, &tenantID, authz.PermTenantViewAudit)
	isPlatformAdmin, _ := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermPlatformViewAudit)

	if !isTenantAdmin && !isPlatformAdmin {
		respondError(w, http.StatusForbidden, "insufficient permissions to view audit logs")
		return
	}

	limit := 50
	offset := 0

	// Create filter
	filter := audit.Filter{
		TenantID: &tenantID,
		Limit:    limit,
		Offset:   offset,
	}

	events, total, err := h.auditRepo.List(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list audit events")
		return
	}

	// Map to response
	respEvents := make([]AuditEventResponse, len(events))
	for i, e := range events {
		respEvents[i] = mapAuditEvent(e)
	}

	respondJSON(w, http.StatusOK, ListAuditEventsResponse{
		Events: respEvents,
		Total:  total,
	})
}

// ListPlatformAuditEvents lists platform-level audit events
// @Summary List platform audit events
// @Description List platform-level audit events (requires platform admin)
// @Tags Audit
// @Produce json
// @Success 200 {object} ListAuditEventsResponse
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit [get]
// @Security SessionCookie
func (h *Handler) ListPlatformAuditEvents(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())

	// Authorization: Platform admin only
	isPlatformAdmin, _ := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermPlatformViewAudit)
	if !isPlatformAdmin {
		respondError(w, http.StatusForbidden, "platform admin access required")
		return
	}

	// Platform logs have empty tenant_id (nil filter implies ANY, empty string implies System?)
	// Our repo implementation: if TenantID string is "", it checks IS NULL.
	// So we pass pointer to empty string.
	emptyTenant := ""
	limit := 50
	offset := 0

	filter := audit.Filter{
		TenantID: &emptyTenant,
		Limit:    limit,
		Offset:   offset,
	}

	events, total, err := h.auditRepo.List(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list audit events")
		return
	}

	// Map to response
	respEvents := make([]AuditEventResponse, len(events))
	for i, e := range events {
		respEvents[i] = mapAuditEvent(e)
	}

	respondJSON(w, http.StatusOK, ListAuditEventsResponse{
		Events: respEvents,
		Total:  total,
	})
}

func mapAuditEvent(e audit.Event) AuditEventResponse {
	return AuditEventResponse{
		ID:        e.ID,
		Type:      e.Type,
		TenantID:  e.TenantID,
		ActorID:   e.ActorID,
		Resource:  e.Resource,
		IPAddress: e.IPAddress,
		UserAgent: e.UserAgent,
		Metadata:  e.Metadata,
		CreatedAt: e.Timestamp.Format(time.RFC3339),
	}
}

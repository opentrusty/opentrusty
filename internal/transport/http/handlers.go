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

// @title OpenTrusty API
// @version DOCS_VERSION_PLACEHOLDER
// @description Production-grade Identity Provider. [🏠 Back to Documentation Home](../index.html)
// @termsOfService https://opentrusty.org/terms/

// @contact.name OpenTrusty Support
// @contact.url https://github.com/opentrusty/opentrusty
// @contact.email support@opentrusty.org

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host api.opentrusty.org
// @BasePath /api/v1

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name session_id

package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/identity"
	"github.com/opentrusty/opentrusty/internal/oauth2"
	"github.com/opentrusty/opentrusty/internal/observability/logger"
	"github.com/opentrusty/opentrusty/internal/oidc"
	"github.com/opentrusty/opentrusty/internal/session"
	"github.com/opentrusty/opentrusty/internal/tenant"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Common JSON response keys
const (
	JSONKeyUserID  = "user_id"
	JSONKeyRole    = "role"
	JSONKeyStatus  = "status"
	JSONKeyEmail   = "email"
	JSONKeySession = "session_id"
)

// Handler holds HTTP handlers and dependencies

type Handler struct {
	identityService *identity.Service
	sessionService  *session.Service
	oauth2Service   *oauth2.Service
	authzService    *authz.Service
	tenantService   *tenant.Service
	oidcService     *oidc.Service
	auditLogger     audit.Logger
	// Configuration
	sessionConfig SessionConfig
	mode          string // "auth", "admin", or "all"
}

// SessionConfig holds session cookie configuration
type SessionConfig struct {
	CookieName     string
	CookieDomain   string
	CookiePath     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite http.SameSite
}

// NewHandler creates a new HTTP handler
func NewHandler(
	identSvc *identity.Service,
	sessSvc *session.Service,
	oauthSvc *oauth2.Service,
	authzSvc *authz.Service,
	tenantSvc *tenant.Service,
	oidcSvc *oidc.Service,
	auditLogger audit.Logger,
	sessConfig SessionConfig,
	mode string,
) *Handler {
	return &Handler{
		identityService: identSvc,
		sessionService:  sessSvc,
		oauth2Service:   oauthSvc,
		authzService:    authzSvc,
		tenantService:   tenantSvc,
		oidcService:     oidcSvc,
		auditLogger:     auditLogger,
		sessionConfig:   sessConfig,
		mode:            mode,
	}
}

// NewRouter creates a new HTTP router
func NewRouter(h *Handler, rateLimiter *RateLimiter, mode string) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(RateLimitMiddleware(rateLimiter))
	r.Use(func(handler http.Handler) http.Handler {
		return otelhttp.NewHandler(handler, "http_request",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)
	})
	r.Use(LoggingMiddleware())
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check (Available in all modes)
	r.Get("/health", h.HealthCheck)

	// Auth Mode: Top-level Routes (OIDC, OAuth2)
	if mode == "auth" || mode == "all" {
		// OIDC Discovery & JWKS (Phase II.2)
		r.Get("/.well-known/openid-configuration", h.Discovery)
		r.Get("/jwks.json", h.JWKS)

		// OAuth2 routes (Tenant-Scoped)
		r.Route("/oauth2", func(r chi.Router) {
			r.Use(TenantMiddleware)
			r.With(h.AuthMiddleware).Get("/authorize", h.Authorize)
			r.Post("/token", h.Token)
			r.Post("/revoke", h.Revoke)
		})
	}

	// Consolidate /api/v1 routes to avoid double-mount panic in "all" mode
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(TenantMiddleware)

		// Auth Plane Endpoints
		if mode == "auth" || mode == "all" {
			r.Group(func(r chi.Router) {
				r.Use(h.CSRFMiddleware) // Enforce CSRF protection for Auth Plane (Login/Logout)
				r.Post("/auth/login", h.Login)
				r.Post("/auth/logout", h.Logout)
				// Note: /auth/register is DISABLED but would be here
				r.Post("/auth/register", h.Register)
			})
		}

		// Admin Plane Endpoints
		if mode == "admin" || mode == "all" {
			// Session Check (Required for Console)
			// Available in Admin mode so Console can check if cookie is valid
			r.With(h.AuthMiddleware).Get("/auth/me", h.GetCurrentUser)

			// Protected Admin routes
			r.Group(func(r chi.Router) {
				r.Use(h.AuthMiddleware)
				r.Use(h.CSRFMiddleware) // Enforce CSRF protection for Admin Plane

				// User profile (Self) - Available in Admin for "My Profile" page
				r.Get("/user/profile", h.GetProfile)
				r.Put("/user/profile", h.UpdateProfile)
				r.Post("/user/change-password", h.ChangePassword)

				// Tenant management (Platform & Tenant assignments)
				r.Route("/tenants", func(r chi.Router) {
					// List/Create tenants are Platform-level actions
					r.Get("/", h.ListTenants)
					r.Post("/", h.CreateTenant)

					// Specific tenant operations
					r.Route("/{tenantID}", func(r chi.Router) {
						r.Use(func(next http.Handler) http.Handler {
							return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								tid := chi.URLParam(r, "tenantID")
								ctx := context.WithValue(r.Context(), tenantIDKey, tid)
								next.ServeHTTP(w, r.WithContext(ctx))
							})
						})
						r.Route("/users/{userID}/roles", func(r chi.Router) {
							r.Post("/", h.AssignTenantRole)
							r.Delete("/{role}", h.RevokeTenantRole)
						})
						// OAuth2 Client Management
						r.Route("/clients", func(r chi.Router) {
							r.Get("/", h.ListClients)
							r.Post("/", h.RegisterClient)
							r.Route("/{clientID}", func(r chi.Router) {
								r.Get("/", h.GetClient)
								r.Delete("/", h.DeleteClient)
								r.Post("/secret", h.RegenerateClientSecret)
							})
						})
					})
				})
			})
		}
	})

	return r
}

// HealthResponse represents the health status response (RFC draft-inadarei-api-health-check)
type HealthResponse struct {
	Status  string `json:"status" example:"pass"`
	Service string `json:"service" example:"opentrusty"`
	Version string `json:"version,omitempty" example:"v1.0.0"`
}

// HealthCheck returns the health status
// @Summary Health Check
// @Description Checks if the service is up and running. Returns "pass", "fail", or "warn".
// @Tags System
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Set headers per best practices
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	respondJSON(w, http.StatusOK, HealthResponse{
		Status:  "pass",
		Service: "opentrusty",
	})
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email      string `json:"email" binding:"required" example:"user@example.com"`
	Password   string `json:"password" binding:"required" example:"secret123"`
	GivenName  string `json:"given_name" example:"John"`
	FamilyName string `json:"family_name" example:"Doe"`
}

// Register handles user registration
// @Summary Register a new user (DISABLED)
// @Description Anonymous registration is disabled for security; admins must be provisioned by platform admins
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration Data"
// @Success 201 {object} map[string]any
// @Failure 403 {object} map[string]string
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Anonymous registration is disabled in the control plane model.
	// Admin accounts must be provisioned by existing platform admins.
	// See: docs/architecture/tenant-context-resolution.md (Anonymous Registration Status)

	slog.WarnContext(r.Context(), "anonymous registration attempt blocked",
		"ip_address", getIPAddress(r),
		"user_agent", r.UserAgent(),
	)

	respondError(w, http.StatusForbidden, "anonymous registration is disabled; admin accounts must be provisioned by platform administrators")
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

// Login handles user login
// @Summary Login
// @Description Authenticate admin user and create a session (tenant derived from user record)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credentials"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string "non-admin user"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// Security Hardening: Reject tenant context from client
	// Per docs/architecture/tenant-context-resolution.md Rule D
	if r.Header.Get("X-Tenant-ID") != "" || r.URL.Query().Get("tenant_id") != "" {
		slog.WarnContext(r.Context(), "tenant context spoofing attempt on /auth/login",
			"ip_address", getIPAddress(r),
			"user_agent", r.UserAgent(),
		)
		respondError(w, http.StatusBadRequest, "tenant context must not be provided; derived from user record post-authentication")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Control Plane Login Model (Rule D):
	// - Authenticate by email + password globally (no tenant filtering)
	// - Derive tenant_id from authenticated user record
	// - Only allow admin-capable roles

	// Use Authenticate with empty string tenant for global lookup
	user, err := h.identityService.Authenticate(r.Context(), "", req.Email, req.Password)
	if err != nil {
		h.auditLogger.Log(r.Context(), audit.Event{
			Type:     audit.TypeLoginFailed,
			Resource: req.Email,
			Metadata: map[string]any{audit.AttrReason: "invalid_credentials"},
		})
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Extract tenant for audit and permission checks
	userTenantID := ""
	if user.TenantID != nil {
		userTenantID = *user.TenantID
	}

	// Control Plane Authorization: Only admin-capable users can login to UI
	// Members authenticate via OAuth2 flows to external applications
	hasAdminRole := false

	// Check platform-level admin permissions
	isPlatformAdmin, err := h.authzService.HasPermission(r.Context(), user.ID, authz.ScopePlatform, nil, authz.PermPlatformManageTenants)
	if err == nil && isPlatformAdmin {
		hasAdminRole = true
	}

	// Check tenant-level admin permissions
	if !hasAdminRole && userTenantID != "" {
		isTenantAdmin, err := h.authzService.HasPermission(r.Context(), user.ID, authz.ScopeTenant, &userTenantID, authz.PermTenantManageUsers)
		if err == nil && isTenantAdmin {
			hasAdminRole = true
		}
	}

	if !hasAdminRole {
		h.auditLogger.Log(r.Context(), audit.Event{
			Type:     audit.TypeLoginFailed,
			TenantID: userTenantID,
			ActorID:  user.ID,
			Resource: req.Email,
			Metadata: map[string]any{audit.AttrReason: "insufficient_privileges"},
		})
		respondError(w, http.StatusForbidden, "access denied: admin role required for UI login")
		return
	}

	// Session Rotation (Hardening Step): Destroy old session if it exists
	if oldSessionID := h.getSessionFromCookie(r); oldSessionID != "" {
		_ = h.sessionService.Destroy(r.Context(), oldSessionID)
	}

	// Determine session namespace based on mode
	// If in "all" or "admin" mode, we treat this as an admin session
	namespace := "admin"
	if h.mode == "auth" {
		namespace = "auth"
	}

	// Create session with immutable tenant_id from user record
	sess, err := h.sessionService.Create(
		r.Context(),
		user.TenantID,
		user.ID,
		getIPAddress(r),
		r.UserAgent(),
		namespace,
	)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to create session", logger.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	h.setSessionCookie(w, sess.ID)

	h.auditLogger.Log(r.Context(), audit.Event{
		Type:      audit.TypeLoginSuccess,
		TenantID:  userTenantID,
		ActorID:   user.ID,
		Resource:  audit.ResourceSession,
		IPAddress: getIPAddress(r),
		UserAgent: r.UserAgent(),
		Metadata:  map[string]any{audit.AttrSessionID: sess.ID},
	})

	respondJSON(w, http.StatusOK, map[string]any{
		JSONKeyUserID: user.ID,
		JSONKeyEmail:  user.Email,
	})
}

// Logout handles user logout
// @Summary Logout
// @Description Destroy the current session
// @Tags Auth
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID" example("tenant_12345")
// @Security CookieAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := h.getSessionFromCookie(r)
	if sessionID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	sess, err := h.sessionService.Get(r.Context(), sessionID)
	if err == nil {
		sessionTenant := ""
		if sess.TenantID != nil {
			sessionTenant = *sess.TenantID
		}

		h.auditLogger.Log(r.Context(), audit.Event{
			Type:      audit.TypeLogout,
			TenantID:  sessionTenant,
			ActorID:   sess.UserID,
			Resource:  audit.ResourceSession,
			IPAddress: getIPAddress(r),
			UserAgent: r.UserAgent(),
			Metadata:  map[string]any{"session_id": sess.ID},
		})
		h.sessionService.Destroy(r.Context(), sessionID)
	}

	h.clearSessionCookie(w)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

// GetCurrentUser returns the current user's information
// @Summary Get current user
// @Description Get the current authenticated user's information
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
// @Security SessionCookie
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Tenant context is derived from session by AuthMiddleware
	// See: docs/architecture/tenant-context-resolution.md
	userID := GetUserID(r.Context())

	// Authorization Check: PermUserReadProfile required (Self)
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermUserReadProfile)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "read profile access required")
		return
	}

	user, err := h.identityService.GetUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	assignments, _ := h.authzService.GetUserRoleAssignments(r.Context(), userID)

	// Derive tenant context from role assignments
	// Platform admins should NOT have tenant context
	var currentTenant map[string]any
	for _, assignment := range assignments {
		// Only set tenant context if user has tenant-scoped roles and NOT platform admin
		if assignment.Scope == "tenant" && assignment.Context != nil && *assignment.Context != "" {
			// Fetch tenant info to include name
			tenant, err := h.tenantService.GetTenant(r.Context(), *assignment.Context)
			if err == nil {
				currentTenant = map[string]any{
					"tenant_id":   tenant.ID,
					"tenant_name": tenant.Name,
				}
				break // Use first tenant found
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"user_id":        user.ID,
			"email":          user.Email,
			"email_verified": user.EmailVerified,
			"profile":        user.Profile,
		},
		"role_assignments": assignments,
		"current_tenant":   currentTenant, // null for platform admins
	})
}

// GetProfile returns the user's profile
// @Summary Get user profile
// @Description Get the current user's profile information
// @Tags User
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Router /user/profile [get]
// @Security SessionCookie
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Tenant context from session
	userID := GetUserID(r.Context())

	// Authorization Check: PermUserReadProfile required
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermUserReadProfile)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "read profile access required")
		return
	}

	user, err := h.identityService.GetUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"user_id":        user.ID,
		"email":          user.Email,
		"email_verified": user.EmailVerified,
		"profile":        user.Profile,
	})
}

// UpdateProfileRequest represents the profile data for update
type UpdateProfileRequest = identity.Profile

// UpdateProfile updates the user's profile
// @Summary Update user profile
// @Description Update the current user's profile information
// @Tags User
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile Data"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /user/profile [put]
// @Security SessionCookie
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Tenant context from session
	userID := GetUserID(r.Context())

	// Authorization Check: PermUserWriteProfile required
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermUserWriteProfile)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "update profile access required")
		return
	}

	var profile identity.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.identityService.UpdateProfile(r.Context(), userID, profile); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "profile updated successfully",
	})
}

// ChangePasswordRequest represents password change data
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword handles password change requests
// @Summary Change password
// @Description Change the current user's password
// @Tags User
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password Change Data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /user/change-password [post]
// @Security SessionCookie
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Tenant context from session
	userID := GetUserID(r.Context())

	// Authorization Check: PermUserChangePassword required
	allowed, err := h.authzService.HasPermission(r.Context(), userID, authz.ScopePlatform, nil, authz.PermUserChangePassword)
	if err != nil || !allowed {
		respondError(w, http.StatusForbidden, "change password access required")
		return
	}

	var req ChangePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.identityService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		switch err {
		case identity.ErrInvalidCredentials:
			respondError(w, http.StatusUnauthorized, "invalid old password")
		case identity.ErrWeakPassword:
			respondError(w, http.StatusBadRequest, "new password does not meet security requirements")
		default:
			respondError(w, http.StatusInternalServerError, "failed to change password")
		}
		return
	}

	h.auditLogger.Log(r.Context(), audit.Event{
		Type:      audit.TypePasswordChanged,
		TenantID:  GetTenantID(r.Context()),
		ActorID:   userID,
		Resource:  audit.ResourceUserCredentials,
		IPAddress: getIPAddress(r),
		UserAgent: r.UserAgent(),
	})

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "password changed successfully",
	})
}

// Helper functions
func (h *Handler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.sessionConfig.CookieName,
		Value:    sessionID,
		Path:     h.sessionConfig.CookiePath,
		Domain:   h.sessionConfig.CookieDomain,
		Secure:   h.sessionConfig.CookieSecure,
		HttpOnly: h.sessionConfig.CookieHTTPOnly,
		SameSite: h.sessionConfig.CookieSameSite,
		MaxAge:   86400, // 24 hours
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   h.sessionConfig.CookieName,
		Value:  "",
		Path:   h.sessionConfig.CookiePath,
		Domain: h.sessionConfig.CookieDomain,
		MaxAge: -1,
	})
}

func (h *Handler) getSessionFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(h.sessionConfig.CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}

func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

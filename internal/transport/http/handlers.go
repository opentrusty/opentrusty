// @title OpenTrusty API
// @version 1.0.0
// @description Production-grade Identity Provider
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
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

// Handler holds HTTP handlers and dependencies
type Handler struct {
	identityService *identity.Service
	sessionService  *session.Service
	oauth2Service   *oauth2.Service
	authzService    *authz.Service
	tenantService   *tenant.Service
	oidcService     *oidc.Service
	auditLogger     audit.Logger
	sessionConfig   SessionConfig
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
	identityService *identity.Service,
	sessionService *session.Service,
	oauth2Service *oauth2.Service,
	authzService *authz.Service,
	tenantService *tenant.Service,
	oidcService *oidc.Service,
	auditLogger audit.Logger,
	sessionConfig SessionConfig,
) *Handler {
	return &Handler{
		identityService: identityService,
		sessionService:  sessionService,
		oauth2Service:   oauth2Service,
		authzService:    authzService,
		tenantService:   tenantService,
		oidcService:     oidcService,
		auditLogger:     auditLogger,
		sessionConfig:   sessionConfig,
	}
}

// NewRouter creates a new HTTP router
func NewRouter(h *Handler, rateLimiter *RateLimiter) *chi.Mux {
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

	// Health check
	r.Get("/health", h.HealthCheck)

	// OIDC Discovery & JWKS (Phase II.2)
	// RFC OIDC Discovery Section 4
	r.Get("/.well-known/openid-configuration", h.Discovery)
	r.Get("/jwks.json", h.JWKS)

	// OAuth2 routes (Tenant-Scoped)
	r.Route("/oauth2", func(r chi.Router) {
		r.Use(TenantMiddleware)
		r.Use(RequireTenant)

		// Authorize endpoint requires user authentication (session)
		// RFC 6749 Section 4.1.1
		r.With(h.AuthMiddleware).Get("/authorize", h.Authorize)

		// Token endpoint uses client authentication
		// RFC 6749 Section 4.1.3
		r.Post("/token", h.Token)

		// Revoke endpoint (RFC 7009)
		r.Post("/revoke", h.Revoke)
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public (Tenant-Agnostic) Endpoints
		// None currently in this group.

		// Tenant-Scoped Endpoints (FAIL-CLOSED)
		r.Group(func(r chi.Router) {
			r.Use(TenantMiddleware)

			// Authentication (Tenant-Scoped)
			r.Use(RequireTenant)
			r.Post("/auth/register", h.Register)
			r.Post("/auth/login", h.Login)
			r.Post("/auth/logout", h.Logout)

			// Protected routes
			r.Group(func(r chi.Router) {
				r.Use(h.AuthMiddleware)

				// Get current user
				r.Get("/auth/me", h.GetCurrentUser)

				// User profile
				r.Get("/user/profile", h.GetProfile)
				r.Put("/user/profile", h.UpdateProfile)
				r.Post("/user/change-password", h.ChangePassword)

				// Tenant management (Platform & Tenant assignments)
				r.Route("/tenants", func(r chi.Router) {
					// List/Create tenants are Platform-level actions (Contextual tenant optional)
					r.Post("/", h.CreateTenant)

					// Specific tenant operations require identification
					r.Route("/{tenantID}", func(r chi.Router) {
						r.Use(func(next http.Handler) http.Handler {
							return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								tid := chi.URLParam(r, "tenantID")
								ctx := context.WithValue(r.Context(), "tenant_id", tid)
								next.ServeHTTP(w, r.WithContext(ctx))
							})
						})
						r.Post("/users", h.ProvisionTenantUser)
						r.Get("/users", h.ListTenantUsers)
						r.Route("/users/{userID}/roles", func(r chi.Router) {
							r.Post("/", h.AssignTenantRole)
							r.Delete("/{role}", h.RevokeTenantRole)
						})
						// OAuth2 Client Management
						r.Post("/oauth2/clients", h.RegisterClient)
					})
				})
			})
		})
	})

	return r
}

// HealthCheck returns the health status
// @Summary Health Check
// @Description Checks if the service is up and running
// @Tags System
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "opentrusty",
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
// @Summary Register a new user
// @Description Register a new user in the current tenant
// @Tags Auth
// @Accept json
// @Produce json
// @Param tenant_id header string true "Tenant ID"
// @Param request body RegisterRequest true "Registration Data"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	profile := identity.Profile{
		GivenName:  req.GivenName,
		FamilyName: req.FamilyName,
		FullName:   req.GivenName + " " + req.FamilyName,
	}

	tenantID := r.Context().Value("tenant_id").(string)
	user, err := h.identityService.ProvisionIdentity(r.Context(), tenantID, req.Email, profile)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to provision user",
			logger.Error(err),
			logger.Email(req.Email),
		)

		switch err {
		case identity.ErrUserAlreadyExists:
			respondError(w, http.StatusConflict, "user already exists")
		case identity.ErrInvalidEmail:
			respondError(w, http.StatusBadRequest, "invalid email address")
		default:
			respondError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	// 2. Set password
	if err := h.identityService.AddPassword(r.Context(), user.ID, req.Password); err != nil {
		slog.ErrorContext(r.Context(), "failed to set password",
			logger.Error(err),
			"user_id", user.ID,
		)
		// TODO: Systematic cleanup if identity exists but password failed?
		// For now, identity exists but passwordless.
		respondError(w, http.StatusBadRequest, "failed to set password: "+err.Error())
		return
	}

	h.auditLogger.Log(r.Context(), audit.Event{
		Type:      audit.TypeUserCreated,
		TenantID:  tenantID,
		ActorID:   user.ID, // Self-registration
		Resource:  "user",
		IPAddress: getIPAddress(r),
		Metadata:  map[string]any{"email": user.Email},
	})

	respondJSON(w, http.StatusCreated, map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

// Login handles user login
// @Summary Login
// @Description Authenticate user and create a session
// @Tags Auth
// @Accept json
// @Produce json
// @Param tenant_id header string true "Tenant ID"
// @Param request body LoginRequest true "Credentials"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Context().Value("tenant_id").(string)
	user, err := h.identityService.Authenticate(r.Context(), tenantID, req.Email, req.Password)
	if err != nil {
		h.auditLogger.Log(r.Context(), audit.Event{
			Type:      audit.TypeLoginFailed,
			TenantID:  tenantID,
			Resource:  req.Email,
			IPAddress: getIPAddress(r),
			UserAgent: r.UserAgent(),
			Metadata:  map[string]any{"reason": "invalid_credentials"},
		})
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Create session
	sess, err := h.sessionService.Create(
		r.Context(),
		user.TenantID,
		user.ID,
		getIPAddress(r),
		r.UserAgent(),
	)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to create session", logger.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Set session cookie
	h.setSessionCookie(w, sess.ID)

	h.auditLogger.Log(r.Context(), audit.Event{
		Type:      audit.TypeLoginSuccess,
		TenantID:  tenantID,
		ActorID:   user.ID,
		Resource:  "session",
		IPAddress: getIPAddress(r),
		UserAgent: r.UserAgent(),
		Metadata:  map[string]any{"session_id": sess.ID},
	})

	respondJSON(w, http.StatusOK, map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

// Logout handles user logout
// @Summary Logout
// @Description Destroy the current session
// @Tags Auth
// @Produce json
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
		h.auditLogger.Log(r.Context(), audit.Event{
			Type:      audit.TypeLogout,
			TenantID:  sess.TenantID,
			ActorID:   sess.UserID,
			Resource:  "session",
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

// GetCurrentUser returns the current authenticated user identity
// @Summary Get Current User
// @Description Retrieve details of the currently logged-in user
// @Tags User
// @Produce json
// @Security CookieAuth
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]string
// @Router /auth/me [get]
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

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

// GetProfile returns the user profile
// @Summary Get User Profile
// @Description Retrieve the profile of the current user
// @Tags User
// @Produce json
// @Security CookieAuth
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]string
// @Router /user/profile [get]
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

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

// UpdateProfile updates the user profile
// @Summary Update Profile
// @Description Update the profile information
// @Tags User
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body identity.Profile true "New Profile"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /user/profile [put]
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

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

// ChangePassword changes the user password
// @Summary Change Password
// @Description Update the password for the current user
// @Tags User
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body ChangePasswordRequest true "Password Change Data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /user/change-password [post]
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req ChangePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.identityService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
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
		TenantID:  r.Context().Value("tenant_id").(string),
		ActorID:   userID,
		Resource:  "user_credentials",
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

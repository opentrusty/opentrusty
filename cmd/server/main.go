package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/config"
	"github.com/opentrusty/opentrusty/internal/identity"
	"github.com/opentrusty/opentrusty/internal/oauth2"
	"github.com/opentrusty/opentrusty/internal/observability/logger"
	"github.com/opentrusty/opentrusty/internal/observability/metrics"
	"github.com/opentrusty/opentrusty/internal/observability/tracing"
	"github.com/opentrusty/opentrusty/internal/oidc"
	"github.com/opentrusty/opentrusty/internal/session"
	"github.com/opentrusty/opentrusty/internal/store/postgres"
	"github.com/opentrusty/opentrusty/internal/tenant"
	transportHTTP "github.com/opentrusty/opentrusty/internal/transport/http"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.InitLogger(logger.Config{
		Level:       cfg.Observability.LogLevel,
		Format:      cfg.Observability.LogFormat,
		ServiceName: cfg.Observability.ServiceName,
	})
	slog.Info("starting opentrusty identity provider")

	// Initialize context
	ctx := context.Background()

	// Initialize tracer
	tracer, err := tracing.New(ctx, tracing.Config{
		Enabled:        cfg.Observability.OTELEnabled,
		ServiceName:    cfg.Observability.ServiceName,
		ServiceVersion: cfg.Observability.ServiceVersion,
		SamplingRate:   1.0,
	})
	if err != nil {
		slog.Error("failed to initialize tracer", logger.Error(err))
	}
	defer tracer.Shutdown(ctx)

	// Initialize meter
	_, err = metrics.New(ctx, metrics.Config{
		Enabled: cfg.Observability.OTELEnabled,
	}, cfg.Observability.ServiceName)
	if err != nil {
		slog.Error("failed to initialize meter", logger.Error(err))
	}

	// Initialize database
	db, err := postgres.New(ctx, postgres.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Database:     cfg.Database.Database,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	})
	if err != nil {
		slog.Error("failed to connect to database", logger.Error(err))
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	storeSessionRepo := postgres.NewSessionRepository(db)
	projectRepo := postgres.NewProjectRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	userProjectRoleRepo := postgres.NewUserProjectRoleRepository(db)
	clientRepo := postgres.NewClientRepository(db)
	codeRepo := postgres.NewAuthorizationCodeRepository(db)
	accessRepo := postgres.NewAccessTokenRepository(db)
	refreshRepo := postgres.NewRefreshTokenRepository(db)
	tenantRepo := postgres.NewTenantRepository(db)
	tenantRoleRepo := postgres.NewTenantRoleRepository(db)

	// Initialize helpers
	auditLogger := audit.NewSlogLogger()
	passwordHasher := identity.NewPasswordHasher(
		cfg.Security.Argon2Memory,
		cfg.Security.Argon2Iterations,
		cfg.Security.Argon2Parallelism,
		cfg.Security.Argon2SaltLength,
		cfg.Security.Argon2KeyLength,
	)

	// Initialize services
	identityService := identity.NewService(
		userRepo,
		passwordHasher,
		auditLogger,
		cfg.Security.LockoutMaxAttempts,
		cfg.Security.LockoutDuration,
	)
	sessionService := session.NewService(storeSessionRepo, cfg.Session.Lifetime, cfg.Session.IdleTimeout)

	// Phase II.1: Initialize OIDC Service
	oidcService, err := oidc.NewService("http://localhost:8080") // TODO: Configurable issuer
	if err != nil {
		slog.Error("failed to initialize OIDC service", logger.Error(err))
		os.Exit(1)
	}

	oauth2Service := oauth2.NewService(
		clientRepo,
		codeRepo,
		accessRepo,
		refreshRepo,
		auditLogger,
		oidcService,
	)
	authzService := authz.NewService(projectRepo, roleRepo, userProjectRoleRepo)
	tenantService := tenant.NewService(tenantRepo, tenantRoleRepo, auditLogger)

	// Rate Limiter
	rateLimiter := transportHTTP.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)

	// Configure SameSite mode
	sameSite := http.SameSiteLaxMode
	switch cfg.Session.CookieSameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "None":
		sameSite = http.SameSiteNoneMode
	}

	// Initialize HTTP handler
	handler := transportHTTP.NewHandler(
		identityService,
		sessionService,
		oauth2Service,
		authzService,
		tenantService,
		oidcService,
		auditLogger,
		transportHTTP.SessionConfig{
			CookieName:     cfg.Session.CookieName,
			CookieDomain:   cfg.Session.CookieDomain,
			CookiePath:     cfg.Session.CookiePath,
			CookieSecure:   cfg.Session.CookieSecure,
			CookieHTTPOnly: cfg.Session.CookieHTTPOnly,
			CookieSameSite: sameSite,
		},
	)

	// Create router
	router := transportHTTP.NewRouter(handler, rateLimiter)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start session cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := sessionService.CleanupExpired(ctx); err != nil {
				slog.ErrorContext(ctx, "failed to cleanup expired sessions", logger.Error(err))
			}
		}
	}()

	// Start server
	go func() {
		slog.Info("starting http server", logger.Component("server"), logger.Operation("listen"))
		slog.Info(fmt.Sprintf("listening on %s", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", logger.Error(err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", logger.Error(err))
	}

	slog.Info("server stopped")
}

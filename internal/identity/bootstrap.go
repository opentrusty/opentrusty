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

package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/authz"
	"github.com/opentrusty/opentrusty/internal/id"
	"github.com/opentrusty/opentrusty/internal/rbac"
)

const (
	EnvBootstrapAdminEmail    = "OT_BOOTSTRAP_ADMIN_EMAIL"
	EnvBootstrapAdminTenantID = "OT_BOOTSTRAP_ADMIN_TENANT_ID"
)

// BootstrapService manages the initial initialization of the system
type BootstrapService struct {
	identityService *Service
	authzRepo       authz.AssignmentRepository
	roleRepo        authz.RoleRepository
	auditLogger     audit.Logger
}

// NewBootstrapService creates a new bootstrap service
func NewBootstrapService(
	identityService *Service,
	authzRepo authz.AssignmentRepository,
	roleRepo authz.RoleRepository,
	auditLogger audit.Logger,
) *BootstrapService {
	return &BootstrapService{
		identityService: identityService,
		authzRepo:       authzRepo,
		roleRepo:        roleRepo,
		auditLogger:     auditLogger,
	}
}

// Bootstrap checks for bootstrap configuration and executes it if necessary
func (s *BootstrapService) Bootstrap(ctx context.Context) error {
	email := os.Getenv(EnvBootstrapAdminEmail)
	tenantID := os.Getenv(EnvBootstrapAdminTenantID)

	if email == "" {
		return nil
	}

	// 1. Check if ANY platform admin already exists
	roleID := rbac.RoleIDPlatformAdmin
	exists, err := s.authzRepo.CheckExists(roleID, authz.ScopePlatform, nil)
	if err != nil {
		return fmt.Errorf("failed to check for existing platform admin: %w", err)
	}

	if exists {
		// Already bootstrapped or admin exists, skip silently
		return nil
	}

	// 2. Look up the user by email and tenant
	var tID *string
	if tenantID != "" {
		tID = &tenantID
	}
	user, err := s.identityService.repo.GetByEmail(tID, email)
	if err != nil {
		// User not found, create it
		fmt.Printf("Bootstrap user not found, creating new platform admin: %s\n", email)

		// Generate random password
		pwBytes := make([]byte, 16)
		if _, randErr := rand.Read(pwBytes); randErr != nil {
			return fmt.Errorf("failed to generate random password: %w", randErr)
		}
		password := base64.RawURLEncoding.EncodeToString(pwBytes) + "!" // Ensure complexity often requires special char depending on rules, adding '!' just in case

		// Provision Identity
		profile := Profile{
			GivenName:  "Platform",
			FamilyName: "Admin",
			FullName:   "Platform Admin",
		}
		// We call ProvisionIdentity on Service? ProvisionIdentity expects Profile struct which is in identity package.
		// bootstrap.go is in identity package, so Profile is available.
		user, err = s.identityService.ProvisionIdentity(ctx, tenantID, email, profile)
		if err != nil {
			return fmt.Errorf("failed to provision bootstrap user: %w", err)
		}

		// Set Password
		if err := s.identityService.AddPassword(ctx, user.ID, password); err != nil {
			return fmt.Errorf("failed to set bootstrap user password: %w", err)
		}

		fmt.Printf("\n\n=== BOOTSTRAP ADMIN CREDENTIALS ===\nEmail: %s\nPassword: %s\n===================================\n\n", email, password)
	}

	// 3. Assign the platform admin role
	assignment := &authz.Assignment{
		ID:             id.NewUUIDv7(),
		UserID:         user.ID,
		RoleID:         roleID,
		Scope:          authz.ScopePlatform,
		ScopeContextID: nil,
		GrantedAt:      time.Now(),
		GrantedBy:      audit.ActorSystemBootstrap,
	}

	if err := s.authzRepo.Grant(assignment); err != nil {
		return fmt.Errorf("failed to grant platform admin role during bootstrap: %w", err)
	}

	// 4. Record audit log
	s.auditLogger.Log(ctx, audit.Event{
		Type:     audit.TypePlatformAdminBootstrap,
		TenantID: tenantID,
		ActorID:  user.ID,
		Resource: audit.ResourcePlatform,
		Metadata: map[string]any{
			audit.AttrEmail:    email,
			audit.AttrTenantID: tenantID,
			audit.AttrRoleID:   roleID,
		},
	})

	fmt.Printf("Successfully bootstrapped initial Platform Admin: %s (Tenant: %s)", email, tenantID)
	return nil
}

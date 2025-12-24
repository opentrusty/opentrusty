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

//go:build integration
// +build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/opentrusty/opentrusty/internal/identity"
)

// TestPurpose: Validates that the database repository maintains strict tenant isolation, preventing cross-tenant data leakage during user retrieval by email.
// Scope: Database Integration Test
// Security: Multi-tenant Data Separation (CWE-284)
// Expected: A user in Tenant A cannot be retrieved using Tenant B's context, even if they share the same email.
// Test Case ID: ISO-01
// Metadata:
//   - Category: Tenant
//   - Priority: High
//   - Tags: multi-tenancy, security, data-isolation
func TestUserRepository_TenantIsolation(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Use docker-compose defaults if no URL provided
		dbURL = "host=localhost port=5432 user=opentrusty password=opentrusty_dev_password dbname=opentrusty sslmode=disable"
	}

	ctx := context.Background()
	cfg := Config{
		Host:         "localhost",
		Port:         "5432",
		User:         "opentrusty",
		Password:     "opentrusty_dev_password",
		Database:     "opentrusty",
		SSLMode:      "disable",
		MaxOpenConns: 5,
		MaxIdleConns: 5,
	}

	db, err := New(ctx, cfg)
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	tenantA := "tenant-a"
	tenantB := "tenant-b"
	email := "shared@example.com"

	userA := &identity.User{
		ID:       "user-a",
		TenantID: tenantA,
		Email:    email,
	}

	userB := &identity.User{
		ID:       "user-b",
		TenantID: tenantB,
		Email:    email,
	}

	// 1. Create User A in Tenant A
	err = repo.Create(userA)
	if err != nil {
		t.Fatalf("failed to create user A: %v", err)
	}
	defer repo.db.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userA.ID)

	// 2. Create User B in Tenant B
	err = repo.Create(userB)
	if err != nil {
		t.Fatalf("failed to create user B: %v", err)
	}
	defer repo.db.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userB.ID)

	// 3. Try to get User B using Tenant A context -> Should fail
	_, err = repo.GetByEmail(tenantA, email)
	if err != nil && err != identity.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound or found user A, got error: %v", err)
	}

	// Verify it's actually User A if found
	foundA, _ := repo.GetByEmail(tenantA, email)
	if foundA != nil && foundA.ID != userA.ID {
		t.Errorf("cross-tenant leakage! expected user A, got %s", foundA.ID)
	}

	// 4. Get User B using Tenant B context -> Should succeed
	foundB, err := repo.GetByEmail(tenantB, email)
	if err != nil {
		t.Errorf("failed to get user B in tenant B: %v", err)
	}
	if foundB == nil || foundB.ID != userB.ID {
		t.Errorf("expected user B, got %v", foundB)
	}
}

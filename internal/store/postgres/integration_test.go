//go:build integration
// +build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/opentrusty/opentrusty/internal/identity"
)

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

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
	"testing"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
)

// MockUserRepository is a simple in-memory implementation of UserRepository
type MockUserRepository struct {
	users       map[string]*User
	credentials map[string]*Credentials
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:       make(map[string]*User),
		credentials: make(map[string]*Credentials),
	}
}

func (m *MockUserRepository) Create(user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) AddCredentials(credentials *Credentials) error {
	m.credentials[credentials.UserID] = credentials
	return nil
}

func (m *MockUserRepository) GetByID(id string) (*User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *MockUserRepository) GetByEmail(tenantID *string, email string) (*User, error) {
	for _, u := range m.users {
		uTenant := ""
		if u.TenantID != nil {
			uTenant = *u.TenantID
		}
		tID := ""
		if tenantID != nil {
			tID = *tenantID
		}

		if uTenant == tID && u.Email == email {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MockUserRepository) Update(user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdateLockout(userID string, failedAttempts int, lockedUntil *time.Time) error {
	u, ok := m.users[userID]
	if !ok {
		return ErrUserNotFound
	}
	u.FailedLoginAttempts = failedAttempts
	u.LockedUntil = lockedUntil
	return nil
}

func (m *MockUserRepository) Delete(id string) error {
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) GetCredentials(userID string) (*Credentials, error) {
	c, ok := m.credentials[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return c, nil
}

func (m *MockUserRepository) UpdatePassword(userID string, passwordHash string) error {
	c, ok := m.credentials[userID]
	if !ok {
		return ErrUserNotFound
	}
	c.PasswordHash = passwordHash
	return nil
}

// TestPurpose: Validates the user authentication flow, including success, failure, and account lockout after multiple failed attempts.
// Scope: Unit Test
// Security: Authentication mechanisms and Brute-force protection (lockout)
// Expected: Successful login for correct credentials, error for wrong credentials, and account lockout after context threshold.
// Test Case ID: IDN-01
func TestIdentity_Service_Authenticate(t *testing.T) {
	repo := NewMockUserRepository()
	hasher := NewPasswordHasher(65536, 3, 4, 16, 32)
	auditLogger := audit.NewSlogLogger()
	s := NewService(repo, hasher, auditLogger, 3, 5*time.Minute)

	ctx := context.Background()
	tenantID := "tenant-1"
	email := "test@example.com"
	password := "SecurePassword123"

	// 1. Provision user
	user, err := s.ProvisionIdentity(ctx, tenantID, email, Profile{FullName: "Test User"})
	if err != nil {
		t.Fatalf("failed to provision: %v", err)
	}

	// 2. Add password
	err = s.AddPassword(ctx, user.ID, password)
	if err != nil {
		t.Fatalf("failed to add password: %v", err)
	}

	// 3. Success authentication
	authSet, err := s.Authenticate(ctx, tenantID, email, password)
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if authSet.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, authSet.ID)
	}

	// 4. Failed authentication (wrong password)
	_, err = s.Authenticate(ctx, tenantID, email, "WrongPassword")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	// 5. Account lockout
	s.Authenticate(ctx, tenantID, email, "WrongPassword")          // Total failed: 2
	_, err = s.Authenticate(ctx, tenantID, email, "WrongPassword") // Total failed: 3 (Threshold met)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for 3rd failed attempt, got %v", err)
	}

	// 4th attempt should be locked out
	_, err = s.Authenticate(ctx, tenantID, email, password)
	if err != ErrAccountLocked {
		t.Errorf("expected ErrAccountLocked, got %v", err)
	}
}

// TestPurpose: Validates that provisioning an identity fails if a user with the same email already exists in the same tenant.
// Scope: Unit Test
// Security: Data Integrity and Unique Constraint Enforcement
// Expected: ErrUserAlreadyExists when email is already registered in the same tenant.
// Test Case ID: IDN-02
func TestIdentity_Service_ProvisionIdentity_Conflict(t *testing.T) {
	repo := NewMockUserRepository()
	hasher := NewPasswordHasher(65536, 3, 4, 16, 32)
	s := NewService(repo, hasher, audit.NewSlogLogger(), 3, 5*time.Minute)

	ctx := context.Background()
	tenantID := "tenant-1"
	email := "conflict@example.com"

	s.ProvisionIdentity(ctx, tenantID, email, Profile{})
	_, err := s.ProvisionIdentity(ctx, tenantID, email, Profile{})
	if err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

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
	"strings"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
	"github.com/opentrusty/opentrusty/internal/id"
	"golang.org/x/crypto/argon2"
)

// PasswordHasher handles password hashing using Argon2id
type PasswordHasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// NewPasswordHasher creates a new password hasher with Argon2id
func NewPasswordHasher(memory, iterations uint32, parallelism uint8, saltLength, keyLength uint32) *PasswordHasher {
	return &PasswordHasher{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  saltLength,
		keyLength:   keyLength,
	}
}

// Hash hashes a password using Argon2id
func (h *PasswordHasher) Hash(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.iterations,
		h.memory,
		h.parallelism,
		h.keyLength,
	)

	// Encode as: $argon2id$v=19$m=memory,t=iterations,p=parallelism$salt$hash
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.memory,
		h.iterations,
		h.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// Verify verifies a password against a hash
func (h *PasswordHasher) Verify(password, encodedHash string) (bool, error) {
	// Parse the encoded hash format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	// Split by $ - format produces: ["argon2id", "v=19", "m=65536,t=3,p=4", "salt", "hash"]
	parts := []byte(encodedHash)
	var sections []string
	start := 0
	for i, c := range parts {
		if c == '$' {
			if i > start {
				sections = append(sections, string(parts[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(parts) {
		sections = append(sections, string(parts[start:]))
	}

	// Expected 5 sections: ["argon2id", "v=19", "m=65536,t=3,p=4", "salt", "hash"]
	if len(sections) != 5 || sections[0] != "argon2id" {
		return false, fmt.Errorf("invalid hash format: got %d sections", len(sections))
	}

	// Parse version
	var version int
	if _, err := fmt.Sscanf(sections[1], "v=%d", &version); err != nil {
		return false, fmt.Errorf("invalid version: %w", err)
	}

	// Parse parameters
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(sections[2], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("invalid parameters: %w", err)
	}

	saltB64 := sections[3]
	hashB64 := sections[4]

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Hash the password with the same parameters
	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(expectedHash)),
	)

	// Compare hashes using constant-time comparison
	if len(actualHash) != len(expectedHash) {
		return false, nil
	}

	var diff byte
	for i := range actualHash {
		diff |= actualHash[i] ^ expectedHash[i]
	}

	return diff == 0, nil
}

// Service provides identity-related business logic
type Service struct {
	repo               UserRepository
	hasher             *PasswordHasher
	auditLogger        audit.Logger
	lockoutMaxAttempts int
	lockoutDuration    time.Duration
}

// NewService creates a new identity service
func NewService(
	repo UserRepository,
	hasher *PasswordHasher,
	auditLogger audit.Logger,
	lockoutMaxAttempts int,
	lockoutDuration time.Duration,
) *Service {
	return &Service{
		repo:               repo,
		hasher:             hasher,
		auditLogger:        auditLogger,
		lockoutMaxAttempts: lockoutMaxAttempts,
		lockoutDuration:    lockoutDuration,
	}
}

// ProvisionIdentity creates a new user identity without credentials
func (s *Service) ProvisionIdentity(ctx context.Context, email string, profile Profile) (*User, error) {
	// Validate email
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	// Check if user already exists
	existing, err := s.repo.GetByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Create user
	if profile.Picture == "" {
		profile.Picture = GenerateRandomAvatar(email)
	}
	if profile.Nickname == "" {
		// Use email prefix as nickname if not provided
		parts := strings.Split(email, "@")
		if len(parts) > 0 {
			profile.Nickname = parts[0]
		}
	}

	user := &User{
		ID:            id.NewUUIDv7(),
		Email:         email,
		EmailVerified: false,
		Profile:       profile,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	return user, nil
}

// AddPassword adds a password credential to an existing user
func (s *Service) AddPassword(ctx context.Context, userID, password string) error {
	// Validate password strength
	if !isStrongPassword(password) {
		return ErrWeakPassword
	}

	// Hash password
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	credentials := &Credentials{
		UserID:       userID,
		PasswordHash: passwordHash,
	}

	if err := s.repo.AddCredentials(credentials); err != nil {
		return fmt.Errorf("failed to add credentials: %w", err)
	}

	return nil
}

// SetPassword sets or updates a user's password without requiring the old password (administrative action)
func (s *Service) SetPassword(ctx context.Context, userID, password string) error {
	// Validate password strength
	if !isStrongPassword(password) {
		return ErrWeakPassword
	}

	// Hash password
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if credentials exist
	_, err = s.repo.GetCredentials(userID)
	if err != nil {
		if err == ErrUserNotFound {
			// Add new credentials
			credentials := &Credentials{
				UserID:       userID,
				PasswordHash: passwordHash,
			}
			return s.repo.AddCredentials(credentials)
		}
		return fmt.Errorf("failed to check existing credentials: %w", err)
	}

	// Update existing credentials
	if err := s.repo.UpdatePassword(userID, passwordHash); err != nil {
		return fmt.Errorf("failed to update credentials: %w", err)
	}

	return nil
}

// Authenticate authenticates a user with email and password
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	// Get user by email globally
	user, err := s.repo.GetByEmail(email)

	if err != nil {
		// Audit failed attempt (unknown user)
		s.auditLogger.Log(ctx, audit.Event{
			Type:     audit.TypeLoginFailed,
			Resource: email,
			Metadata: map[string]any{audit.AttrReason: "user_not_found"},
		})
		return nil, ErrInvalidCredentials
	}

	// Check if locked out
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		s.auditLogger.Log(ctx, audit.Event{
			Type:     audit.TypeLoginFailed,
			ActorID:  user.ID,
			Resource: "login",
			Metadata: map[string]any{audit.AttrReason: "locked_out"},
		})
		return nil, ErrAccountLocked
	}

	// Get credentials
	credentials, err := s.repo.GetCredentials(user.ID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	valid, err := s.hasher.Verify(password, credentials.PasswordHash)
	if err != nil || !valid {
		// Increment failed attempts
		newAttempts := user.FailedLoginAttempts + 1
		var newLockedUntil *time.Time

		if newAttempts >= s.lockoutMaxAttempts {
			until := time.Now().Add(s.lockoutDuration)
			newLockedUntil = &until
			// Audit lockout
			s.auditLogger.Log(ctx, audit.Event{
				Type:     audit.TypeUserLocked,
				ActorID:  user.ID,
				Resource: "login",
				Metadata: map[string]any{audit.AttrAttempts: newAttempts},
			})
		}

		// Update lockout status
		_ = s.repo.UpdateLockout(user.ID, newAttempts, newLockedUntil)

		// Audit failed attempt
		s.auditLogger.Log(ctx, audit.Event{
			Type:     audit.TypeLoginFailed,
			ActorID:  user.ID,
			Resource: "login",
			Metadata: map[string]any{
				audit.AttrReason:   "invalid_password",
				audit.AttrAttempts: newAttempts,
			},
		})

		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts if > 0
	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		_ = s.repo.UpdateLockout(user.ID, 0, nil)
	}

	// Audit success
	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeLoginSuccess,
		ActorID:    user.ID,
		Resource:   "login",
		TargetID:   user.ID,
		TargetName: user.Email,
	})

	return user, nil
}

// GetByEmail retrieves a user by email globally
func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(email)
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, userID string) (*User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile updates user profile information
func (s *Service) UpdateProfile(ctx context.Context, userID string, profile Profile) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	user.Profile = profile
	return s.repo.Update(user)
}

// ChangePassword changes user password
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Get credentials
	credentials, err := s.repo.GetCredentials(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify old password
	valid, err := s.hasher.Verify(oldPassword, credentials.PasswordHash)
	if err != nil || !valid {
		return ErrInvalidCredentials
	}

	// Validate new password
	if !isStrongPassword(newPassword) {
		return ErrWeakPassword
	}

	// Hash new password
	newHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.UpdatePassword(userID, newHash)
}

// Helper functions
func isValidEmail(email string) bool {
	// Basic email validation
	// In production, use a proper email validation library
	return len(email) > 3 && len(email) < 255
}

func isStrongPassword(password string) bool {
	// Password must be at least 8 characters
	// In production, implement more sophisticated password strength checking
	return len(password) >= 8
}

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

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty/internal/identity"
)

// UserRepository implements identity.UserRepository
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user identity
func (r *UserRepository) Create(user *identity.User) error {
	ctx := context.Background()
	now := time.Now()
	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO users (
			id, tenant_id, email, email_verified,
			given_name, family_name, full_name, nickname, picture, locale, timezone,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		user.ID, user.TenantID, user.Email, user.EmailVerified,
		user.Profile.GivenName, user.Profile.FamilyName, user.Profile.FullName,
		user.Profile.Nickname, user.Profile.Picture, user.Profile.Locale, user.Profile.Timezone,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	user.CreatedAt = now
	user.UpdatedAt = now

	return nil
}

// AddCredentials adds credentials for a user
func (r *UserRepository) AddCredentials(credentials *identity.Credentials) error {
	ctx := context.Background()
	now := time.Now()

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO credentials (user_id, password_hash, updated_at)
		VALUES ($1, $2, $3)
	`, credentials.UserID, credentials.PasswordHash, now)
	if err != nil {
		return fmt.Errorf("failed to insert credentials: %w", err)
	}

	credentials.UpdatedAt = now

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*identity.User, error) {
	ctx := context.Background()

	var user identity.User
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, email_verified,
			given_name, family_name, full_name, nickname, picture, locale, timezone,
			created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.EmailVerified,
		&user.Profile.GivenName, &user.Profile.FamilyName, &user.Profile.FullName,
		&user.Profile.Nickname, &user.Profile.Picture, &user.Profile.Locale, &user.Profile.Timezone,
		&user.CreatedAt, &user.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, identity.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(tenantID, email string) (*identity.User, error) {
	ctx := context.Background()

	var user identity.User
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, email_verified,
			given_name, family_name, full_name, nickname, picture, locale, timezone,
			created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
	`, tenantID, email).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.EmailVerified,
		&user.Profile.GivenName, &user.Profile.FamilyName, &user.Profile.FullName,
		&user.Profile.Nickname, &user.Profile.Picture, &user.Profile.Locale, &user.Profile.Timezone,
		&user.CreatedAt, &user.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, identity.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

// Update updates user information
func (r *UserRepository) Update(user *identity.User) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE users SET
			email = $3,
			email_verified = $4,
			given_name = $5,
			family_name = $6,
			full_name = $7,
			nickname = $8,
			picture = $9,
			locale = $10,
			timezone = $11
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`,
		user.ID, user.TenantID, user.Email, user.EmailVerified,
		user.Profile.GivenName, user.Profile.FamilyName, user.Profile.FullName,
		user.Profile.Nickname, user.Profile.Picture, user.Profile.Locale, user.Profile.Timezone,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return identity.ErrUserNotFound
	}

	return nil
}

// UpdateLockout updates user lockout status
func (r *UserRepository) UpdateLockout(userID string, failedAttempts int, lockedUntil *time.Time) error {
	query := `
		UPDATE users
		SET failed_login_attempts = $1, locked_until = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.pool.Exec(context.Background(), query, failedAttempts, lockedUntil, userID)
	if err != nil {
		return fmt.Errorf("failed to update user lockout status: %w", err)
	}
	return nil
}

// Delete soft-deletes a user
func (r *UserRepository) Delete(id string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE users SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, id, time.Now())

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return identity.ErrUserNotFound
	}

	return nil
}

// GetCredentials retrieves user credentials
func (r *UserRepository) GetCredentials(userID string) (*identity.Credentials, error) {
	ctx := context.Background()

	var creds identity.Credentials

	err := r.db.pool.QueryRow(ctx, `
		SELECT user_id, password_hash, updated_at
		FROM credentials
		WHERE user_id = $1
	`, userID).Scan(&creds.UserID, &creds.PasswordHash, &creds.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, identity.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return &creds, nil
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(userID string, passwordHash string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE credentials SET password_hash = $2
		WHERE user_id = $1
	`, userID, passwordHash)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return identity.ErrUserNotFound
	}

	return nil
}

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
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty/internal/session"
)

// SessionRepository implements session.Repository
type SessionRepository struct {
	db *DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(sess *session.Session) error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO sessions (id, tenant_id, user_id, ip_address, user_agent, expires_at, created_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		sess.ID, sess.TenantID, sess.UserID, sess.IPAddress, sess.UserAgent,
		sess.ExpiresAt, sess.CreatedAt, sess.LastSeenAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// Get retrieves a session by ID
func (r *SessionRepository) Get(sessionID string) (*session.Session, error) {
	ctx := context.Background()

	var sess session.Session

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, tenant_id, user_id, ip_address, user_agent, expires_at, created_at, last_seen_at
		FROM sessions
		WHERE id = $1
	`, sessionID).Scan(
		&sess.ID, &sess.TenantID, &sess.UserID, &sess.IPAddress, &sess.UserAgent,
		&sess.ExpiresAt, &sess.CreatedAt, &sess.LastSeenAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, session.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &sess, nil
}

// Update updates session last seen time
func (r *SessionRepository) Update(sess *session.Session) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE sessions SET last_seen_at = $2
		WHERE id = $1
	`, sess.ID, sess.LastSeenAt)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return session.ErrSessionNotFound
	}

	return nil
}

// Delete deletes a session
func (r *SessionRepository) Delete(sessionID string) error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM sessions WHERE id = $1
	`, sessionID)

	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *SessionRepository) DeleteByUserID(userID string) error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM sessions WHERE user_id = $1
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *SessionRepository) DeleteExpired() error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM sessions WHERE expires_at < $1
	`, time.Now())

	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

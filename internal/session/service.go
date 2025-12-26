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

package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// Service provides session management business logic
type Service struct {
	repo        Repository
	lifetime    time.Duration
	idleTimeout time.Duration
}

// NewService creates a new session service
func NewService(repo Repository, lifetime, idleTimeout time.Duration) *Service {
	return &Service{
		repo:        repo,
		lifetime:    lifetime,
		idleTimeout: idleTimeout,
	}
}

// Create creates a new session for a user
func (s *Service) Create(ctx context.Context, tenantID *string, userID, ipAddress, userAgent, namespace string) (*Session, error) {
	session := &Session{
		ID:         generateSessionID(),
		TenantID:   tenantID,
		UserID:     userID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Namespace:  namespace,
		ExpiresAt:  time.Now().Add(s.lifetime),
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}

	if err := s.repo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// Get retrieves and validates a session
func (s *Service) Get(ctx context.Context, sessionID string) (*Session, error) {
	session, err := s.repo.Get(sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// Check if session is expired
	if session.IsExpired() {
		s.repo.Delete(sessionID)
		return nil, ErrSessionExpired
	}

	// Check if session is idle
	if session.IsIdle(s.idleTimeout) {
		s.repo.Delete(sessionID)
		return nil, ErrSessionExpired
	}

	return session, nil
}

// Refresh refreshes a session's last seen time
func (s *Service) Refresh(ctx context.Context, sessionID string) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastSeenAt = time.Now()
	return s.repo.Update(session)
}

// Destroy destroys a session
func (s *Service) Destroy(ctx context.Context, sessionID string) error {
	return s.repo.Delete(sessionID)
}

// DestroyAllForUser destroys all sessions for a user
func (s *Service) DestroyAllForUser(ctx context.Context, userID string) error {
	return s.repo.DeleteByUserID(userID)
}

// CleanupExpired removes all expired sessions
func (s *Service) CleanupExpired(ctx context.Context) error {
	return s.repo.DeleteExpired()
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

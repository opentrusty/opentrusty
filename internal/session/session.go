package session

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionInvalid  = errors.New("session invalid")
)

// Session represents a user session
type Session struct {
	ID         string
	TenantID   string
	UserID     string
	IPAddress  string
	UserAgent  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastSeenAt time.Time
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsIdle checks if the session has been idle for too long
func (s *Session) IsIdle(idleTimeout time.Duration) bool {
	return time.Since(s.LastSeenAt) > idleTimeout
}

// Repository defines the interface for session persistence
type Repository interface {
	// Create creates a new session
	Create(session *Session) error

	// Get retrieves a session by ID
	Get(sessionID string) (*Session, error)

	// Update updates session last seen time
	Update(session *Session) error

	// Delete deletes a session
	Delete(sessionID string) error

	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(userID string) error

	// DeleteExpired deletes all expired sessions
	DeleteExpired() error
}

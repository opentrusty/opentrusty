package identity

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrWeakPassword       = errors.New("password does not meet security requirements")
	ErrAccountLocked      = errors.New("account is locked")
)

// User represents a user identity in the system
type User struct {
	ID                  string
	TenantID            string
	Email               string
	EmailVerified       bool
	Profile             Profile
	FailedLoginAttempts int
	LockedUntil         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

// Profile represents user profile information
type Profile struct {
	GivenName  string
	FamilyName string
	FullName   string
	Nickname   string
	Picture    string
	Locale     string
	Timezone   string
}

// Credentials represents user authentication credentials
type Credentials struct {
	UserID       string
	PasswordHash string
	UpdatedAt    time.Time
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// Create creates a new user identity
	Create(user *User) error

	// AddCredentials adds credentials for a user
	AddCredentials(credentials *Credentials) error

	// GetByID retrieves a user by ID
	GetByID(id string) (*User, error)

	// GetByEmail retrieves a user by email within a tenant
	GetByEmail(tenantID, email string) (*User, error)

	// Update updates user information
	Update(user *User) error

	// UpdateLockout updates user lockout status
	UpdateLockout(userID string, failedAttempts int, lockedUntil *time.Time) error

	// Delete soft-deletes a user
	Delete(id string) error

	// GetCredentials retrieves user credentials
	GetCredentials(userID string) (*Credentials, error)

	// UpdatePassword updates user password
	UpdatePassword(userID string, passwordHash string) error
}

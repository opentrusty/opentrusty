package oauth2

import (
	"context"
	"time"
)

// KeyType represents the type of key
type KeyType string

const (
	KeyTypeRSA KeyType = "RSA"
)

// Algorithm represents the signing algorithm
type Algorithm string

const (
	AlgorithmRS256 Algorithm = "RS256"
)

// Key represents a cryptographic key for signing tokens
type Key struct {
	ID                  string
	Type                KeyType
	Algorithm           Algorithm
	PublicKey           string // PEM encoded or JWK JSON
	PrivateKeyEncrypted []byte // Encrypted private key
	CreatedAt           time.Time
	ExpiresAt           time.Time
}

// KeyRepository defines the interface for key persistence
type KeyRepository interface {
	// Create stores a new key
	Create(ctx context.Context, key *Key) error

	// GetActiveKey retrieves the current active signing key
	GetActiveKey(ctx context.Context) (*Key, error)

	// ListValidKeys retrieves all valid keys (active and not expired)
	ListValidKeys(ctx context.Context) ([]*Key, error)

	// Rotate creates a new key and makes it active
	// Rotate(ctx context.Context) (*Key, error) // For future
}

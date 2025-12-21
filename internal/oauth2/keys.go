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

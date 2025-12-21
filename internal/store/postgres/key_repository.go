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

	"github.com/opentrusty/opentrusty/internal/oauth2"
)

// KeyRepository implements oauth2.KeyRepository
type KeyRepository struct {
	db *DB
}

// NewKeyRepository creates a new key repository
func NewKeyRepository(db *DB) *KeyRepository {
	return &KeyRepository{db: db}
}

// Create stores a new key
func (r *KeyRepository) Create(ctx context.Context, key *oauth2.Key) error {
	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO openid_keys (
			id, type, algorithm, public_key, private_key_encrypted, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		key.ID, key.Type, key.Algorithm, key.PublicKey, key.PrivateKeyEncrypted, key.CreatedAt, key.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	return nil
}

// GetActiveKey retrieves the most recent valid key
func (r *KeyRepository) GetActiveKey(ctx context.Context) (*oauth2.Key, error) {
	var key oauth2.Key
	err := r.db.pool.QueryRow(ctx, `
		SELECT id, type, algorithm, public_key, private_key_encrypted, created_at, expires_at
		FROM openid_keys
		WHERE expires_at > $1
		ORDER BY created_at DESC
		LIMIT 1
	`, time.Now()).Scan(
		&key.ID, &key.Type, &key.Algorithm, &key.PublicKey, &key.PrivateKeyEncrypted, &key.CreatedAt, &key.ExpiresAt,
	)

	if err != nil {
		// return nil, fmt.Errorf("failed to get active key: %w", err)
		// Return specific error or nil if not found, let service handle generation
		return nil, err
	}

	return &key, nil
}

// ListValidKeys retrieves all valid keys
func (r *KeyRepository) ListValidKeys(ctx context.Context) ([]*oauth2.Key, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, type, algorithm, public_key, private_key_encrypted, created_at, expires_at
		FROM openid_keys
		WHERE expires_at > $1
		ORDER BY created_at DESC
	`, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	defer rows.Close()

	var keys []*oauth2.Key
	for rows.Next() {
		var key oauth2.Key
		if err := rows.Scan(&key.ID, &key.Type, &key.Algorithm, &key.PublicKey, &key.PrivateKeyEncrypted, &key.CreatedAt, &key.ExpiresAt); err != nil {
			return nil, fmt.Errorf("failed to scan key: %w", err)
		}
		keys = append(keys, &key)
	}

	return keys, nil
}

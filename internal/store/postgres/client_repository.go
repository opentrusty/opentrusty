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
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty/internal/oauth2"
)

// ClientRepository implements oauth2.ClientRepository
type ClientRepository struct {
	db *DB
}

// NewClientRepository creates a new client repository
func NewClientRepository(db *DB) *ClientRepository {
	return &ClientRepository{db: db}
}

// Create creates a new OAuth2 client
func (r *ClientRepository) Create(client *oauth2.Client) error {
	ctx := context.Background()

	redirectURIs, err := json.Marshal(client.RedirectURIs)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect URIs: %w", err)
	}

	allowedScopes, err := json.Marshal(client.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}

	grantTypes, err := json.Marshal(client.GrantTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal grant types: %w", err)
	}

	responseTypes, err := json.Marshal(client.ResponseTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal response types: %w", err)
	}

	var ownerID sql.NullString
	if client.OwnerID != "" {
		ownerID = sql.NullString{String: client.OwnerID, Valid: true}
	}

	_, err = r.db.pool.Exec(ctx, `
		INSERT INTO oauth2_clients (
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`,
		client.ID, client.ClientID, client.TenantID, client.ClientSecretHash, client.ClientName, client.ClientURI, client.LogoURI,
		redirectURIs, allowedScopes, grantTypes, responseTypes,
		client.TokenEndpointAuthMethod, client.AccessTokenLifetime, client.RefreshTokenLifetime, client.IDTokenLifetime,
		ownerID, client.IsTrusted, client.IsActive, client.CreatedAt, client.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

// GetByClientID retrieves a client by client_id
func (r *ClientRepository) GetByClientID(clientID string) (*oauth2.Client, error) {
	ctx := context.Background()

	var client oauth2.Client
	var redirectURIsJSON, allowedScopesJSON, grantTypesJSON, responseTypesJSON []byte
	var clientURI, logoURI, ownerID sql.NullString
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT 
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at, deleted_at
		FROM oauth2_clients
		WHERE client_id = $1 AND deleted_at IS NULL
	`, clientID).Scan(
		&client.ID, &client.ClientID, &client.TenantID, &client.ClientSecretHash, &client.ClientName, &clientURI, &logoURI,
		&redirectURIsJSON, &allowedScopesJSON, &grantTypesJSON, &responseTypesJSON,
		&client.TokenEndpointAuthMethod, &client.AccessTokenLifetime, &client.RefreshTokenLifetime, &client.IDTokenLifetime,
		&ownerID, &client.IsTrusted, &client.IsActive, &client.CreatedAt, &client.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, oauth2.ErrClientNotFound
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(redirectURIsJSON, &client.RedirectURIs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redirect URIs: %w", err)
	}
	if err := json.Unmarshal(allowedScopesJSON, &client.AllowedScopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
	}
	if err := json.Unmarshal(grantTypesJSON, &client.GrantTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal grant types: %w", err)
	}
	if err := json.Unmarshal(responseTypesJSON, &client.ResponseTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response types: %w", err)
	}

	if clientURI.Valid {
		client.ClientURI = clientURI.String
	}
	if logoURI.Valid {
		client.LogoURI = logoURI.String
	}
	if ownerID.Valid {
		client.OwnerID = ownerID.String
	}
	if deletedAt.Valid {
		client.DeletedAt = &deletedAt.Time
	}

	return &client, nil
}

// GetByID retrieves a client by internal ID
func (r *ClientRepository) GetByID(id string) (*oauth2.Client, error) {
	ctx := context.Background()

	var client oauth2.Client
	var redirectURIsJSON, allowedScopesJSON, grantTypesJSON, responseTypesJSON []byte
	var ownerID sql.NullString
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT 
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at, deleted_at
		FROM oauth2_clients
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&client.ID, &client.ClientID, &client.TenantID, &client.ClientSecretHash, &client.ClientName, &client.ClientURI, &client.LogoURI,
		&redirectURIsJSON, &allowedScopesJSON, &grantTypesJSON, &responseTypesJSON,
		&client.TokenEndpointAuthMethod, &client.AccessTokenLifetime, &client.RefreshTokenLifetime, &client.IDTokenLifetime,
		&ownerID, &client.IsTrusted, &client.IsActive, &client.CreatedAt, &client.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, oauth2.ErrClientNotFound
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(redirectURIsJSON, &client.RedirectURIs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redirect URIs: %w", err)
	}
	if err := json.Unmarshal(allowedScopesJSON, &client.AllowedScopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
	}
	if err := json.Unmarshal(grantTypesJSON, &client.GrantTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal grant types: %w", err)
	}
	if err := json.Unmarshal(responseTypesJSON, &client.ResponseTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response types: %w", err)
	}

	if ownerID.Valid {
		client.OwnerID = ownerID.String
	}
	if deletedAt.Valid {
		client.DeletedAt = &deletedAt.Time
	}

	return &client, nil
}

// Update updates client information
func (r *ClientRepository) Update(client *oauth2.Client) error {
	ctx := context.Background()

	redirectURIs, err := json.Marshal(client.RedirectURIs)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect URIs: %w", err)
	}

	allowedScopes, err := json.Marshal(client.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}

	grantTypes, err := json.Marshal(client.GrantTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal grant types: %w", err)
	}

	responseTypes, err := json.Marshal(client.ResponseTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal response types: %w", err)
	}

	result, err := r.db.pool.Exec(ctx, `
		UPDATE oauth2_clients SET
			client_name = $2,
			client_uri = $3,
			logo_uri = $4,
			redirect_uris = $5,
			allowed_scopes = $6,
			grant_types = $7,
			response_types = $8,
			token_endpoint_auth_method = $9,
			access_token_lifetime = $10,
			refresh_token_lifetime = $11,
			id_token_lifetime = $12,
			is_trusted = $13,
			is_active = $14
		WHERE id = $1 AND deleted_at IS NULL
	`,
		client.ID, client.ClientName, client.ClientURI, client.LogoURI,
		redirectURIs, allowedScopes, grantTypes, responseTypes,
		client.TokenEndpointAuthMethod, client.AccessTokenLifetime, client.RefreshTokenLifetime, client.IDTokenLifetime,
		client.IsTrusted, client.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return oauth2.ErrClientNotFound
	}

	return nil
}

// Delete soft-deletes a client
func (r *ClientRepository) Delete(id string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE oauth2_clients SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, id, time.Now())

	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return oauth2.ErrClientNotFound
	}

	return nil
}

// ListByOwner retrieves all clients for an owner
func (r *ClientRepository) ListByOwner(ownerID string) ([]*oauth2.Client, error) {
	ctx := context.Background()

	rows, err := r.db.pool.Query(ctx, `
		SELECT 
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at, deleted_at
		FROM oauth2_clients
		WHERE owner_id = $1 AND deleted_at IS NULL
	`, ownerID)

	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []*oauth2.Client

	for rows.Next() {
		var client oauth2.Client
		var redirectURIsJSON, allowedScopesJSON, grantTypesJSON, responseTypesJSON []byte
		var ownerID sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&client.ID, &client.ClientID, &client.TenantID, &client.ClientSecretHash, &client.ClientName, &client.ClientURI, &client.LogoURI,
			&redirectURIsJSON, &allowedScopesJSON, &grantTypesJSON, &responseTypesJSON,
			&client.TokenEndpointAuthMethod, &client.AccessTokenLifetime, &client.RefreshTokenLifetime, &client.IDTokenLifetime,
			&ownerID, &client.IsTrusted, &client.IsActive, &client.CreatedAt, &client.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}

		if err := json.Unmarshal(redirectURIsJSON, &client.RedirectURIs); err != nil {
			continue // Skip malformed records
		}
		if err := json.Unmarshal(allowedScopesJSON, &client.AllowedScopes); err != nil {
			continue
		}
		if err := json.Unmarshal(grantTypesJSON, &client.GrantTypes); err != nil {
			continue
		}
		if err := json.Unmarshal(responseTypesJSON, &client.ResponseTypes); err != nil {
			continue
		}

		if ownerID.Valid {
			client.OwnerID = ownerID.String
		}
		if deletedAt.Valid {
			client.DeletedAt = &deletedAt.Time
		}

		clients = append(clients, &client)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return clients, nil
}

// ListByTenant retrieves all clients for a tenant
func (r *ClientRepository) ListByTenant(tenantID string) ([]*oauth2.Client, error) {
	ctx := context.Background()

	rows, err := r.db.pool.Query(ctx, `
		SELECT 
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at, deleted_at
		FROM oauth2_clients
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []*oauth2.Client

	for rows.Next() {
		var client oauth2.Client
		var redirectURIsJSON, allowedScopesJSON, grantTypesJSON, responseTypesJSON []byte
		var ownerID sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&client.ID, &client.ClientID, &client.TenantID, &client.ClientSecretHash, &client.ClientName, &client.ClientURI, &client.LogoURI,
			&redirectURIsJSON, &allowedScopesJSON, &grantTypesJSON, &responseTypesJSON,
			&client.TokenEndpointAuthMethod, &client.AccessTokenLifetime, &client.RefreshTokenLifetime, &client.IDTokenLifetime,
			&ownerID, &client.IsTrusted, &client.IsActive, &client.CreatedAt, &client.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}

		if err := json.Unmarshal(redirectURIsJSON, &client.RedirectURIs); err != nil {
			continue
		}
		if err := json.Unmarshal(allowedScopesJSON, &client.AllowedScopes); err != nil {
			continue
		}
		if err := json.Unmarshal(grantTypesJSON, &client.GrantTypes); err != nil {
			continue
		}
		if err := json.Unmarshal(responseTypesJSON, &client.ResponseTypes); err != nil {
			continue
		}

		if ownerID.Valid {
			client.OwnerID = ownerID.String
		}
		if deletedAt.Valid {
			client.DeletedAt = &deletedAt.Time
		}

		clients = append(clients, &client)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return clients, nil
}

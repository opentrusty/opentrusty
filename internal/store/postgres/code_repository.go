package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty/internal/oauth2"
)

// AuthorizationCodeRepository implements oauth2.AuthorizationCodeRepository
type AuthorizationCodeRepository struct {
	db *DB
}

// NewAuthorizationCodeRepository creates a new authorization code repository
func NewAuthorizationCodeRepository(db *DB) *AuthorizationCodeRepository {
	return &AuthorizationCodeRepository{db: db}
}

// Create creates a new authorization code
func (r *AuthorizationCodeRepository) Create(code *oauth2.AuthorizationCode) error {
	ctx := context.Background()

	var usedAt sql.NullTime
	if code.UsedAt != nil {
		usedAt = sql.NullTime{Time: *code.UsedAt, Valid: true}
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO authorization_codes (
			id, code, client_id, user_id, 
			redirect_uri, scope, state, nonce,
			code_challenge, code_challenge_method,
			expires_at, used_at, is_used, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`,
		code.ID, code.Code, code.ClientID, code.UserID,
		code.RedirectURI, code.Scope, code.State, code.Nonce,
		code.CodeChallenge, code.CodeChallengeMethod,
		code.ExpiresAt, usedAt, code.IsUsed, code.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create authorization code: %w", err)
	}

	return nil
}

// GetByCode retrieves an authorization code
func (r *AuthorizationCodeRepository) GetByCode(codeStr string) (*oauth2.AuthorizationCode, error) {
	ctx := context.Background()

	var code oauth2.AuthorizationCode
	var usedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT 
			id, code, client_id, user_id, 
			redirect_uri, scope, state, nonce,
			code_challenge, code_challenge_method,
			expires_at, used_at, is_used, created_at
		FROM authorization_codes
		WHERE code = $1
	`, codeStr).Scan(
		&code.ID, &code.Code, &code.ClientID, &code.UserID,
		&code.RedirectURI, &code.Scope, &code.State, &code.Nonce,
		&code.CodeChallenge, &code.CodeChallengeMethod,
		&code.ExpiresAt, &usedAt, &code.IsUsed, &code.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, oauth2.ErrCodeNotFound
		}
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	if usedAt.Valid {
		code.UsedAt = &usedAt.Time
	}

	return &code, nil
}

// MarkAsUsed marks the code as used
func (r *AuthorizationCodeRepository) MarkAsUsed(code string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE authorization_codes SET is_used = true, used_at = $2
		WHERE code = $1
	`, code, time.Now())

	if err != nil {
		return fmt.Errorf("failed to mark code as used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return oauth2.ErrCodeNotFound
	}

	return nil
}

// Delete deletes an authorization code
func (r *AuthorizationCodeRepository) Delete(code string) error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM authorization_codes WHERE code = $1
	`, code)

	if err != nil {
		return fmt.Errorf("failed to delete code: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired authorization codes
func (r *AuthorizationCodeRepository) DeleteExpired() error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM authorization_codes WHERE expires_at < $1
	`, time.Now())

	if err != nil {
		return fmt.Errorf("failed to delete expired codes: %w", err)
	}

	return nil
}

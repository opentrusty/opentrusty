package oauth2

import (
	"context"
	"testing"
	"time"

	"github.com/opentrusty/opentrusty/internal/audit"
)

// BenchMockCodeRepo ignores usage checks to allow looping
type BenchMockCodeRepo struct {
	code *AuthorizationCode
}

func (m *BenchMockCodeRepo) Create(code *AuthorizationCode) error { return nil }
func (m *BenchMockCodeRepo) GetByCode(code string) (*AuthorizationCode, error) {
	return m.code, nil
}
func (m *BenchMockCodeRepo) MarkAsUsed(code string) error { return nil }
func (m *BenchMockCodeRepo) Delete(code string) error     { return nil }
func (m *BenchMockCodeRepo) DeleteExpired() error         { return nil }

func BenchmarkService_ExchangeCodeForToken(b *testing.B) {
	// Setup Mocks
	clientRepo := &MockClientRepo{
		clients: map[string]*Client{
			"bench-client": {
				ClientID:            "bench-client",
				ClientSecretHash:    hashClientSecret("bench-secret"),
				RedirectURIs:        []string{"https://app.com/cb"},
				AccessTokenLifetime: 3600,
				IsActive:            true,
			},
		},
	}

	validCode := &AuthorizationCode{
		Code:        "valid-code",
		ClientID:    "bench-client",
		RedirectURI: "https://app.com/cb",
		UserID:      "user-1",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	codeRepo := &BenchMockCodeRepo{code: validCode}
	accessRepo := &MockAccessRepo{}
	refreshRepo := &MockRefreshRepo{}

	svc := &Service{
		clientRepo:  clientRepo,
		codeRepo:    codeRepo,
		accessRepo:  accessRepo,
		refreshRepo: refreshRepo,
		auditLogger: audit.NewSlogLogger(),
	}

	req := &TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     "bench-client",
		ClientSecret: "bench-secret",
		Code:         "valid-code",
		RedirectURI:  "https://app.com/cb",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.ExchangeCodeForToken(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

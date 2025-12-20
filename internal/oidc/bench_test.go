package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func BenchmarkService_GenerateIDToken(b *testing.B) {
	// Setup service with static key
	s := &Service{}
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	s.signingKey = key
	s.issuer = "https://auth.opentrusty.org"
	s.kid = "test-kid-1"

	userID := "user-123"
	tenantID := "tenant-456"
	clientID := "client-789"
	nonce := "noncce-abc"
	accessToken := "access-token-xyz"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.GenerateIDToken(userID, tenantID, clientID, nonce, accessToken)
		if err != nil {
			b.Fatal(err)
		}
	}
}

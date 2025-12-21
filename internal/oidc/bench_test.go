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

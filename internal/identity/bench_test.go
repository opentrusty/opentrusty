package identity

import (
	"testing"
)

func BenchmarkPasswordHasher_Hash(b *testing.B) {
	// RFC 9106 recommended parameters
	hasher := NewPasswordHasher(64*1024, 1, 4, 16, 32)
	password := "correct-horse-battery-staple"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(password) // Hash doesn't take ctx in current sig?
		// Check code: func (h *PasswordHasher) Hash(password string) (string, error)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordHasher_Verify(b *testing.B) {
	hasher := NewPasswordHasher(64*1024, 1, 4, 16, 32)
	password := "correct-horse-battery-staple"
	// ctx := context.Background() // Not needed for Hash/Verify
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valid, err := hasher.Verify(password, hash)
		if err != nil || !valid {
			b.Fatalf("verify failed: %v", err)
		}
	}
}

package audit

import (
	"testing"
)

// TestPurpose: Validates that sensitive keys are correctly identified as secrets to prevent them from being logged in plaintext.
// Scope: Unit Test
// Security: Data Masking and Leakage Prevention (CWE-532)
// Expected: Returns true for keys containing 'password', 'token', 'secret', etc., and false for non-sensitive keys.
// Test Case ID: AUD-01
func TestAudit_IsSecret(t *testing.T) {
	tests := []struct {
		key      string
		isSecret bool
	}{
		{"password", true},
		{"Password", true},
		{"PASSWORD", true},
		{"token", true},
		{"access_token", true},
		{"secret", true},
		{"api_key", true},
		{"hash", true},
		{"password_hash", true},
		{"credential", true},
		{"private_key", true},
		{"user_id", false},
		{"tenant_id", false},
		{"email", false},
		{"status", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := isSecret(tt.key); got != tt.isSecret {
				t.Errorf("isSecret(%q) = %v, want %v", tt.key, got, tt.isSecret)
			}
		})
	}
}

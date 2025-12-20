package oidc

import "fmt"

// Error represents a protocol-level OIDC error (OIDC Core).
// Fix for B-PROTOCOL-01 (Rule 4.2 - Compliance)
type Error struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
	URI         string `json:"error_uri,omitempty"`
	State       string `json:"state,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("oidc error: %s (%s)", e.Code, e.Description)
}

// OIDC Standard Error Codes
const (
	ErrInteractionRequired      = "interaction_required"
	ErrLoginRequired            = "login_required"
	ErrConsentRequired          = "consent_required"
	ErrAccountSelectionRequired = "account_selection_required"
)

// NewError creates a new OIDC protocol error
func NewError(code, description string) *Error {
	return &Error{
		Code:        code,
		Description: description,
	}
}

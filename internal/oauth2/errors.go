package oauth2

import "fmt"

// Error represents a protocol-level OAuth2 error (RFC 6749).
// Fix for B-PROTOCOL-01 (Rule 4.2 - Compliance)
type Error struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
	URI         string `json:"error_uri,omitempty"`
	State       string `json:"state,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("oauth2 error: %s (%s)", e.Code, e.Description)
}

// OAuth2 Standard Error Codes
const (
	ErrInvalidRequest         = "invalid_request"
	ErrInvalidClient          = "invalid_client"
	ErrInvalidGrant           = "invalid_grant"
	ErrUnauthorizedClient     = "unauthorized_client"
	ErrUnsupportedGrantType   = "unsupported_grant_type"
	ErrInvalidScope           = "invalid_scope"
	ErrServerError            = "server_error"
	ErrTemporarilyUnavailable = "temporarily_unavailable"
)

// NewError creates a new protocol error
func NewError(code, description string) *Error {
	return &Error{
		Code:        code,
		Description: description,
	}
}

// WithState attaches a state parameter to the error
func (e *Error) WithState(state string) *Error {
	e.State = state
	return e
}

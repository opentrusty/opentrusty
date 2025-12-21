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

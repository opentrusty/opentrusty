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

package tenant

import (
	"errors"
	"time"
)

// Domain errors (ErrTenantNotFound is defined in repository.go)
var (
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrInvalidTenantName   = errors.New("invalid tenant name")
)

// Tenant represents an isolated environment or customer account
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultTenantID is the ID of the default tenant
const DefaultTenantID = "default"

// Status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// TenantMetrics represents summary statistics for a tenant
type TenantMetrics struct {
	TotalUsers    int `json:"total_users"`
	TotalClients  int `json:"total_clients"`
	AuditCount24h int `json:"audit_count_24h"`
}

// Membership represents a user's membership in a tenant
type Membership struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

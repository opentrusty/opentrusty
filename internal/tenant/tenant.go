package tenant

import (
	"time"
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

package tenant

import "time"

// Tenant Roles
const (
	RoleTenantOwner  = "tenant_owner"
	RoleTenantAdmin  = "tenant_admin"
	RoleTenantMember = "tenant_member"
)

// TenantUserRole represents a user's role assignment in a tenant
type TenantUserRole struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	GrantedAt time.Time `json:"granted_at"`
	GrantedBy string    `json:"granted_by"`
}

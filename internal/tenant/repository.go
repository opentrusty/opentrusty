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
	"context"
	"errors"
)

var (
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role assignment already exists")
)

// Repository defines the interface for tenant storage
type Repository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id string) (*Tenant, error)
	GetByName(ctx context.Context, name string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Tenant, error)
}

// RoleRepository defines the interface for tenant role storage
type RoleRepository interface {
	AssignRole(ctx context.Context, role *TenantUserRole) error
	RevokeRole(ctx context.Context, tenantID, userID, role string) error
	GetUserRoles(ctx context.Context, tenantID, userID string) ([]*TenantUserRole, error)
	GetTenantUsers(ctx context.Context, tenantID string) ([]*TenantUserRole, error)
}

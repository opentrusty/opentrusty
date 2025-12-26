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

package authz

import (
	"context"
	"fmt"
)

// Service provides authorization business logic
type Service struct {
	projectRepo    ProjectRepository
	roleRepo       RoleRepository
	assignmentRepo AssignmentRepository
}

// NewService creates a new authorization service
func NewService(
	projectRepo ProjectRepository,
	roleRepo RoleRepository,
	assignmentRepo AssignmentRepository,
) *Service {
	return &Service{
		projectRepo:    projectRepo,
		roleRepo:       roleRepo,
		assignmentRepo: assignmentRepo,
	}
}

// GetUserRoles retrieves all unique role names for a user across all scopes
func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	assignments, err := s.assignmentRepo.ListForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assignments: %w", err)
	}

	roleMap := make(map[string]bool)
	for _, a := range assignments {
		role, err := s.roleRepo.GetByID(a.RoleID)
		if err != nil {
			continue
		}
		roleMap[role.Name] = true
	}

	roleNames := make([]string, 0, len(roleMap))
	for name := range roleMap {
		roleNames = append(roleNames, name)
	}

	return roleNames, nil
}

// UserRoleAssignment represents a role assigned to a user with scope
type UserRoleAssignment struct {
	RoleName string  `json:"role_name"`
	Scope    string  `json:"scope"`
	Context  *string `json:"context,omitempty"`
}

// GetUserRoleAssignments retrieves all role assignments for a user with details
func (s *Service) GetUserRoleAssignments(ctx context.Context, userID string) ([]UserRoleAssignment, error) {
	assignments, err := s.assignmentRepo.ListForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assignments: %w", err)
	}

	var result []UserRoleAssignment
	for _, a := range assignments {
		role, err := s.roleRepo.GetByID(a.RoleID)
		if err != nil {
			continue
		}
		result = append(result, UserRoleAssignment{
			RoleName: role.Name,
			Scope:    string(a.Scope),
			Context:  a.ScopeContextID,
		})
	}

	return result, nil
}

// GetUserProjects retrieves all projects a user has access to (deprecated/legacy support)
func (s *Service) GetUserProjects(ctx context.Context, userID string) ([]*Project, error) {
	return s.projectRepo.ListByUser(userID)
}

// ProjectInfo represents simplified project information for external systems
type ProjectInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UserInfoClaims represents the claims to be returned in the userinfo endpoint
type UserInfoClaims struct {
	Roles    []string       `json:"roles"`
	Projects []*ProjectInfo `json:"projects"`
}

// BuildUserInfoClaims builds the authorization claims for a user
func (s *Service) BuildUserInfoClaims(ctx context.Context, userID string) (*UserInfoClaims, error) {
	// Get user roles
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Get user projects
	projects, err := s.GetUserProjects(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	// Convert projects to simplified format
	projectInfos := make([]*ProjectInfo, 0, len(projects))
	for _, p := range projects {
		projectInfos = append(projectInfos, &ProjectInfo{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		})
	}

	return &UserInfoClaims{
		Roles:    roles,
		Projects: projectInfos,
	}, nil
}

// HasPermission checks if a user has a specific permission at a scope
func (s *Service) HasPermission(ctx context.Context, userID string, scope Scope, scopeContextID *string, permission string) (bool, error) {
	assignments, err := s.assignmentRepo.ListForUser(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user assignments: %w", err)
	}

	for _, a := range assignments {
		// Scope check: assignment scope must be same as requested, OR assignment is platform scope
		// (Platform admin has all permissions at all scopes? Or explicit?
		// Requirement: "Use scoped authorization (platform scope)".
		// Let's stick to explicit match or platform-to-any if that's the model.
		// For now: exact match or platform scope.
		match := false
		if a.Scope == ScopePlatform {
			match = true
		} else if a.Scope == scope {
			if a.ScopeContextID != nil && scopeContextID != nil && *a.ScopeContextID == *scopeContextID {
				match = true
			} else if a.ScopeContextID == nil && scopeContextID == nil {
				match = true
			}
		}

		if !match {
			continue
		}

		role, err := s.roleRepo.GetByID(a.RoleID)
		if err != nil {
			continue
		}

		if role.HasPermission(permission) {
			return true, nil
		}
	}

	return false, nil
}

package authz

import (
	"context"
	"fmt"
)

// Service provides authorization business logic
type Service struct {
	projectRepo         ProjectRepository
	roleRepo            RoleRepository
	userProjectRoleRepo UserProjectRoleRepository
}

// NewService creates a new authorization service
func NewService(
	projectRepo ProjectRepository,
	roleRepo RoleRepository,
	userProjectRoleRepo UserProjectRoleRepository,
) *Service {
	return &Service{
		projectRepo:         projectRepo,
		roleRepo:            roleRepo,
		userProjectRoleRepo: userProjectRoleRepo,
	}
}

// GetUserRoles retrieves all unique role names for a user across all projects
func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	// Get all projects the user has access to
	projects, err := s.userProjectRoleRepo.GetUserProjects(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	// Collect unique role names across all projects
	roleMap := make(map[string]bool)
	for _, project := range projects {
		roles, err := s.userProjectRoleRepo.GetUserRolesInProject(userID, project.ID)
		if err != nil {
			// Log error but continue with other projects
			continue
		}
		for _, role := range roles {
			roleMap[role.Name] = true
		}
	}

	// Convert map to slice
	roleNames := make([]string, 0, len(roleMap))
	for roleName := range roleMap {
		roleNames = append(roleNames, roleName)
	}

	return roleNames, nil
}

// GetUserProjects retrieves all projects a user has access to
func (s *Service) GetUserProjects(ctx context.Context, userID string) ([]*Project, error) {
	projects, err := s.userProjectRoleRepo.GetUserProjects(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}
	return projects, nil
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

// HasPermission checks if a user has a specific permission in a project
func (s *Service) HasPermission(ctx context.Context, userID, projectID, permission string) (bool, error) {
	// Check if user has access to the project
	hasAccess, err := s.userProjectRoleRepo.HasAccess(userID, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return false, nil
	}

	// Get user's roles in the project
	roles, err := s.userProjectRoleRepo.GetUserRolesInProject(userID, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check if any role has the permission
	for _, role := range roles {
		if role.HasPermission(permission) {
			return true, nil
		}
	}

	return false, nil
}

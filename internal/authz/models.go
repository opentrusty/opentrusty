package authz

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrRoleNotFound         = errors.New("role not found")
	ErrRoleAlreadyExists    = errors.New("role already exists")
	ErrAccessDenied         = errors.New("access denied")
	ErrInvalidPermission    = errors.New("invalid permission")
)

// Project represents a project/resource that users can access
type Project struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Role represents a role with associated permissions
type Role struct {
	ID          string
	Name        string
	Description string
	Permissions []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if p == "*" || p == permission {
			return true
		}
	}
	return false
}

// UserProjectRole represents a user's role assignment in a project
type UserProjectRole struct {
	ID        string
	UserID    string
	ProjectID string
	RoleID    string
	GrantedAt time.Time
	GrantedBy string
}

// ProjectRepository defines the interface for project persistence
type ProjectRepository interface {
	// Create creates a new project
	Create(project *Project) error

	// GetByID retrieves a project by ID
	GetByID(id string) (*Project, error)

	// GetByName retrieves a project by name
	GetByName(name string) (*Project, error)

	// Update updates project information
	Update(project *Project) error

	// Delete soft-deletes a project
	Delete(id string) error

	// ListByOwner retrieves all projects owned by a user
	ListByOwner(ownerID string) ([]*Project, error)

	// ListByUser retrieves all projects a user has access to
	ListByUser(userID string) ([]*Project, error)
}

// RoleRepository defines the interface for role persistence
type RoleRepository interface {
	// Create creates a new role
	Create(role *Role) error

	// GetByID retrieves a role by ID
	GetByID(id string) (*Role, error)

	// GetByName retrieves a role by name
	GetByName(name string) (*Role, error)

	// Update updates role information
	Update(role *Role) error

	// Delete deletes a role
	Delete(id string) error

	// List retrieves all roles
	List() ([]*Role, error)
}

// UserProjectRoleRepository defines the interface for user-project-role persistence
type UserProjectRoleRepository interface {
	// Grant assigns a role to a user in a project
	Grant(upr *UserProjectRole) error

	// Revoke removes a role from a user in a project
	Revoke(userID, projectID, roleID string) error

	// GetUserRolesInProject retrieves all roles a user has in a project
	GetUserRolesInProject(userID, projectID string) ([]*Role, error)

	// GetUserProjects retrieves all projects a user has access to
	GetUserProjects(userID string) ([]*Project, error)

	// GetProjectUsers retrieves all users with access to a project
	GetProjectUsers(projectID string) ([]string, error)

	// HasAccess checks if a user has access to a project
	HasAccess(userID, projectID string) (bool, error)
}

package tenancy

import (
	"context"
	"fmt"
	"net/http"
)

// Role defines a user role in the system.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

// Permission defines a specific permission.
type Permission string

const (
	PermissionRead         Permission = "read"
	PermissionWrite        Permission = "write"
	PermissionDelete       Permission = "delete"
	PermissionManageUsers  Permission = "manage_users"
	PermissionManageRoles  Permission = "manage_roles"
	PermissionManageTenant Permission = "manage_tenant"
)

// RBACEngine implements role-based access control.
type RBACEngine struct {
	rolePermissions map[Role][]Permission
}

// NewRBACEngine creates a new RBAC engine.
func NewRBACEngine() *RBACEngine {
	engine := &RBACEngine{
		rolePermissions: make(map[Role][]Permission),
	}

	// Define default role permissions
	engine.rolePermissions[RoleOwner] = []Permission{
		PermissionRead,
		PermissionWrite,
		PermissionDelete,
		PermissionManageUsers,
		PermissionManageRoles,
		PermissionManageTenant,
	}

	engine.rolePermissions[RoleAdmin] = []Permission{
		PermissionRead,
		PermissionWrite,
		PermissionDelete,
		PermissionManageUsers,
	}

	engine.rolePermissions[RoleMember] = []Permission{
		PermissionRead,
		PermissionWrite,
	}

	engine.rolePermissions[RoleViewer] = []Permission{
		PermissionRead,
	}

	return engine
}

// HasPermission checks if a role has a specific permission.
func (e *RBACEngine) HasPermission(role Role, permission Permission) bool {
	permissions, exists := e.rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// GetRolePermissions returns all permissions for a role.
func (e *RBACEngine) GetRolePermissions(role Role) []Permission {
	return e.rolePermissions[role]
}

// ValidateAccess validates if a user has permission for a resource.
func (e *RBACEngine) ValidateAccess(ctx context.Context, userID string, tenantID string, permission Permission) error {
	// Get user's role in tenant
	role, err := e.getUserRole(ctx, userID, tenantID)
	if err != nil {
		return fmt.Errorf("get user role: %w", err)
	}

	if !e.HasPermission(role, permission) {
		return fmt.Errorf("permission denied: %s", permission)
	}

	return nil
}

// getUserRole gets user's role in a tenant.
func (e *RBACEngine) getUserRole(ctx context.Context, userID, tenantID string) (Role, error) {
	// Placeholder: would query database
	// For now, return member role
	return RoleMember, nil
}

// CheckPermission middleware checks if user has required permission.
func CheckPermission(rbac *RBACEngine, permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := GetTenantFromContext(r.Context())
			if !ok {
				http.Error(w, "No tenant context", http.StatusBadRequest)
				return
			}

			userID := extractUserID(r)
			if userID == "" {
				http.Error(w, "No user context", http.StatusUnauthorized)
				return
			}

			if err := rbac.ValidateAccess(r.Context(), userID, tenant.ID, permission); err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IsValidRole checks if a role string is valid.
func IsValidRole(role string) bool {
	validRoles := map[string]bool{
		string(RoleOwner):  true,
		string(RoleAdmin):  true,
		string(RoleMember): true,
		string(RoleViewer): true,
	}
	return validRoles[role]
}

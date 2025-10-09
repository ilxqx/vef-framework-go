package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/security"
)

// RBACPermissionChecker implements role-based access control (RBAC) permission checking.
// It delegates role permissions loading to a RolePermissionsLoader implementation.
type RBACPermissionChecker struct {
	loader security.RolePermissionsLoader
}

// NewRBACPermissionChecker creates a new RBAC permission checker.
// loader: The strategy for loading role permissions.
func NewRBACPermissionChecker(loader security.RolePermissionsLoader) security.PermissionChecker {
	return &RBACPermissionChecker{
		loader: loader,
	}
}

// HasPermission checks if the principal has the required permission based on their roles.
// System principals always have all permissions.
// For user and external app principals, it loads role permissions sequentially.
// Sequential loading is more efficient for typical use cases (1-3 roles per user).
func (c *RBACPermissionChecker) HasPermission(
	ctx context.Context,
	principal *security.Principal,
	permissionToken string,
) (bool, error) {
	if principal == nil {
		return false, nil
	}

	// If principal has no roles, they have no permissions
	if len(principal.Roles) == 0 {
		return false, nil
	}

	// Load permissions for each role and check if any contains the required permission
	// Using sequential loading for efficiency since most users have only 1-3 roles
	for _, role := range principal.Roles {
		permissions, err := c.loader.LoadPermissions(ctx, role)
		if err != nil {
			return false, err
		}

		// O(1) lookup: check if the permission token exists in the map
		if _, exists := permissions[permissionToken]; exists {
			return true, nil
		}
	}

	return false, nil
}

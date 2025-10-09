package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/security"
)

// RBACDataPermResolver implements role-based data permission resolution.
// It delegates role permissions loading to a RolePermissionsLoader implementation.
type RBACDataPermResolver struct {
	loader security.RolePermissionsLoader
}

// NewRBACDataPermResolver creates a new RBAC data permission resolver.
// loader: The strategy for loading role permissions.
func NewRBACDataPermResolver(loader security.RolePermissionsLoader) security.DataPermissionResolver {
	return &RBACDataPermResolver{
		loader: loader,
	}
}

// ResolveDataScope resolves the applicable DataScope for the given principal and permission token.
// It loads permissions for all roles and returns the first matching DataScope.
// If multiple roles have the same permission token, the first match is returned.
// Returns nil if no matching permission is found.
func (r *RBACDataPermResolver) ResolveDataScope(
	ctx context.Context,
	principal *security.Principal,
	permToken string,
) (security.DataScope, error) {
	if principal == nil {
		return nil, nil
	}

	// If principal has no roles, they have no permissions
	if len(principal.Roles) == 0 {
		return nil, nil
	}

	// Load permissions for each role and find the matching permission
	// Using sequential loading for efficiency since most users have only 1-3 roles
	for _, role := range principal.Roles {
		permissions, err := r.loader.LoadPermissions(ctx, role)
		if err != nil {
			return nil, err
		}

		// O(1) lookup: check if the permission token exists and return its DataScope
		if dataScope, exists := permissions[permToken]; exists {
			return dataScope, nil
		}
	}

	// No matching permission found
	return nil, nil
}

package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/security"
)

// RbacDataPermissionResolver implements role-based data permission resolution.
// It delegates role permissions loading to a RolePermissionsLoader implementation.
type RbacDataPermissionResolver struct {
	loader security.RolePermissionsLoader
}

// NewRbacDataPermissionResolver creates a new RBAC data permission resolver.
// loader: The strategy for loading role permissions.
func NewRbacDataPermissionResolver(loader security.RolePermissionsLoader) security.DataPermissionResolver {
	return &RbacDataPermissionResolver{
		loader: loader,
	}
}

// ResolveDataScope resolves the applicable DataScope for the given principal and permission token.
// When a user has multiple roles with the same permission token but different data scopes,
// this method collects all matching scopes and returns the one with the highest priority.
// Returns nil if no matching permission is found.
func (r *RbacDataPermissionResolver) ResolveDataScope(
	ctx context.Context,
	principal *security.Principal,
	permToken string,
) (security.DataScope, error) {
	// If loader is nil, no data scope resolution is possible
	if r.loader == nil {
		return nil, nil
	}

	// If principal is nil or has no roles, they have no permissions
	if principal == nil || len(principal.Roles) == 0 {
		return nil, nil
	}

	var (
		selectedScope security.DataScope
		maxPriority   = -1
	)

	// Load permissions for each role and collect all matching DataScopes
	// Using sequential loading for efficiency since most users have only 1-3 roles
	for _, role := range principal.Roles {
		permissions, err := r.loader.LoadPermissions(ctx, role)
		if err != nil {
			return nil, err
		}

		if dataScope, exists := permissions[permToken]; exists {
			// Compare priorities and select the scope with the highest priority
			if priority := dataScope.Priority(); priority > maxPriority {
				maxPriority = priority
				selectedScope = dataScope
			}
		}
	}

	// Return the DataScope with the highest priority, or nil if none found
	return selectedScope, nil
}

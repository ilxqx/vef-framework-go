package security

import (
	"context"
	"sync"

	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/set"
)

// RBACPermissionChecker implements role-based access control (RBAC) permission checking.
// It delegates role permission loading to a RolePermissionsLoader implementation.
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
// For user and external app principals, it loads role permissions concurrently.
func (c *RBACPermissionChecker) HasPermission(
	ctx context.Context,
	principal *security.Principal,
	permissionToken string,
) (bool, error) {
	if principal == nil {
		return false, nil
	}

	// System principal has all permissions
	if principal.Type == security.PrincipalTypeSystem {
		return true, nil
	}

	// If principal has no roles, they have no permissions
	if len(principal.Roles) == 0 {
		return false, nil
	}

	// Load permissions for all roles concurrently
	type roleResult struct {
		permissions set.Set[string]
		err         error
	}

	results := make(chan roleResult, len(principal.Roles))

	var wg sync.WaitGroup

	for _, role := range principal.Roles {
		wg.Add(1)

		go func(r string) {
			defer wg.Done()

			permissions, err := c.loader.LoadPermissions(ctx, r)
			results <- roleResult{permissions: permissions, err: err}
		}(role)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Check results
	for result := range results {
		if result.err != nil {
			return false, result.err
		}

		if result.permissions.Contains(permissionToken) {
			return true, nil
		}
	}

	return false, nil
}

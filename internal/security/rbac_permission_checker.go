package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/security"
)

type RbacPermissionChecker struct {
	loader security.RolePermissionsLoader
}

func NewRbacPermissionChecker(loader security.RolePermissionsLoader) security.PermissionChecker {
	return &RbacPermissionChecker{
		loader: loader,
	}
}

// HasPermission uses sequential role loading rather than parallel to optimize for common case (1-3 roles).
func (c *RbacPermissionChecker) HasPermission(
	ctx context.Context,
	principal *security.Principal,
	permissionToken string,
) (bool, error) {
	if principal == nil {
		return false, nil
	}

	if len(principal.Roles) == 0 {
		return false, nil
	}

	for _, role := range principal.Roles {
		permissions, err := c.loader.LoadPermissions(ctx, role)
		if err != nil {
			return false, err
		}

		if _, exists := permissions[permissionToken]; exists {
			return true, nil
		}
	}

	return false, nil
}

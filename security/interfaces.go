package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/orm"
)

type Authenticator interface {
	Supports(authType string) bool
	Authenticate(ctx context.Context, authentication Authentication) (*Principal, error)
}

type TokenGenerator interface {
	Generate(principal *Principal) (*AuthTokens, error)
}

type AuthManager interface {
	Authenticate(ctx context.Context, authentication Authentication) (*Principal, error)
}

type UserLoader interface {
	LoadByUsername(ctx context.Context, username string) (*Principal, string, error)
	LoadById(ctx context.Context, id string) (*Principal, error)
}

type ExternalAppLoader interface {
	LoadById(ctx context.Context, id string) (*Principal, string, error)
}

type PasswordDecryptor interface {
	Decrypt(encryptedPassword string) (string, error)
}

type PermissionChecker interface {
	HasPermission(ctx context.Context, principal *Principal, permToken string) (bool, error)
}

type RolePermissionsLoader interface {
	LoadPermissions(ctx context.Context, role string) (map[string]DataScope, error)
}

type UserInfoLoader interface {
	LoadUserInfo(ctx context.Context, principal *Principal, params map[string]any) (*UserInfo, error)
}

type DataScope interface {
	Key() string
	Priority() int
	Supports(principal *Principal, table *orm.Table) bool
	Apply(principal *Principal, query orm.SelectQuery) error
}

type DataPermissionResolver interface {
	ResolveDataScope(ctx context.Context, principal *Principal, permToken string) (DataScope, error)
}

type DataPermissionApplier interface {
	Apply(query orm.SelectQuery) error
}

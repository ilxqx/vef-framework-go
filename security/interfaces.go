package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/orm"
)

// Authenticator defines the interface for authentication providers.
// Each authenticator supports specific authentication types and validates credentials.
type Authenticator interface {
	// Supports checks if this authenticator can handle the given authentication type.
	Supports(authType string) bool
	// Authenticate validates the provided authentication information and returns a Principal.
	// Returns an error if authentication fails or the authenticator doesn't support the type.
	Authenticate(authentication Authentication) (*Principal, error)
}

// TokenGenerator defines the interface for generating authentication tokens.
// Different implementations can support various token types (JWT, opaque tokens, etc.).
type TokenGenerator interface {
	// Generate creates authentication tokens for the given principal.
	// Returns both access and refresh tokens, or an error if generation fails.
	Generate(principal *Principal) (*AuthTokens, error)
}

// AuthManager is the main entry point for authentication operations.
// It manages multiple authenticators and provides a unified authentication interface.
type AuthManager interface {
	// Authenticate attempts to authenticate the provided authentication information.
	// It delegates to the appropriate authenticator based on the authentication type.
	// Returns a Principal if authentication succeeds, or an error if it fails.
	Authenticate(authentication Authentication) (*Principal, error)
}

// UserLoader defines a strategy for loading user information by username.
// Users of the framework should implement this interface and provide it via the fx container.
// The returned passwordHash should be the hashed password stored in the system.
type UserLoader interface {
	// LoadByUsername loads a user by username and returns the associated Principal and password hash.
	// If the username does not exist, return (nil, "", nil) and the authenticator will treat it as invalid credentials.
	// If an internal error occurs, return a non-nil error.
	LoadByUsername(username string) (*Principal, string, error)
	// LoadById loads a user by id and returns the associated Principal.
	// If the user does not exist, return nil.
	// If an internal error occurs, return a non-nil error.
	LoadById(id string) (*Principal, error)
}

// ExternalAppLoader defines a strategy for loading external application information by appId.
// Users should implement this interface and provide it via the fx container.
// The returned secret should be the shared secret used for HMAC verification.
type ExternalAppLoader interface {
	// LoadById loads an external application by appId and returns the associated Principal and app secret.
	// If the app does not exist, return (nil, "", nil) and the authenticator will treat it as invalid credentials.
	// If an internal error occurs, return a non-nil error.
	LoadById(id string) (*Principal, string, error)
}

// PasswordDecryptor defines the interface for decrypting passwords received from clients.
// Different implementations can support various encryption algorithms (AES, RSA, SM2, SM4, etc.).
// If no decryptor is provided, passwords are assumed to be plaintext.
// Users should implement this interface based on their specific encryption requirements.
type PasswordDecryptor interface {
	// Decrypt decrypts the encrypted password string and returns the plaintext password.
	// The encryptedPassword parameter is typically a base64-encoded or hex-encoded string.
	// Returns an error if decryption fails (e.g., invalid format, wrong key, corrupted data).
	Decrypt(encryptedPassword string) (string, error)
}

// PermissionChecker defines the interface for checking whether a principal has a specific permission.
// Users should implement this interface and provide it via the fx container.
// The framework provides a default RBAC implementation, but users can implement custom logic.
type PermissionChecker interface {
	// HasPermission checks if the given principal has the specified permission.
	HasPermission(ctx context.Context, principal *Principal, permToken string) (bool, error)
}

// RolePermissionsLoader defines a strategy for loading all permissions associated with a role.
// It returns a map where keys are permission tokens and values are their corresponding data scopes.
// This interface is used by the RBAC PermissionChecker and DataPermissionResolver implementations.
// Users should implement this interface to define how role permissions are loaded.
type RolePermissionsLoader interface {
	// LoadPermissions loads all permissions associated with the given role.
	// Returns a map of permission token to DataScope, allowing O(1) permission checks.
	LoadPermissions(ctx context.Context, role string) (map[string]DataScope, error)
}

// DataScope represents an abstract data permission scope that defines access boundaries.
// Each DataScope implementation encapsulates a specific data access pattern (e.g., department-level, organization-level).
// Implementations should be stateless and thread-safe, as they may be shared across multiple requests.
type DataScope interface {
	// Key returns the unique identifier of this data scope.
	Key() string
	// Supports determines whether this data scope is applicable to the given table structure.
	// It checks if the table has the necessary fields required by this scope.
	Supports(principal *Principal, table *orm.Table) bool
	// Apply applies the data permission filter conditions using the provided SelectQuery.
	// This method should use the SelectQuery to add filtering conditions.
	Apply(principal *Principal, query orm.SelectQuery) error
}

// DataPermissionResolver resolves the applicable DataScope instance for a given principal and permission.
type DataPermissionResolver interface {
	// ResolveDataScope loads the DataScope instance applicable to the principal.
	ResolveDataScope(ctx context.Context, principal *Principal, permToken string) (DataScope, error)
}

// DataPermissionApplier applies data permission filtering to queries.
// Thread Safety:
// Instances are NOT required to be thread-safe as they are request-scoped.
type DataPermissionApplier interface {
	// Apply applies data permission filter conditions to the query.
	Apply(query orm.SelectQuery) error
}

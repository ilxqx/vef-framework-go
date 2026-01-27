package security

import (
	"context"
	"time"

	"github.com/ilxqx/vef-framework-go/orm"
)

// Authenticator validates credentials and returns a Principal on success.
// Multiple authenticators can be registered to support different authentication methods
// (e.g., JWT, password, OAuth). Each authenticator declares which authentication
// kinds it supports via the Supports method.
type Authenticator interface {
	// Supports returns true if this authenticator can handle the given authentication kind.
	Supports(kind string) bool
	// Authenticate validates the credentials and returns the authenticated Principal.
	Authenticate(ctx context.Context, authentication Authentication) (*Principal, error)
}

// TokenGenerator creates access and refresh tokens for an authenticated Principal.
// Used after successful authentication to issue JWT tokens or similar credentials.
type TokenGenerator interface {
	// Generate creates a new token pair for the given Principal.
	Generate(principal *Principal) (*AuthTokens, error)
}

// AuthManager orchestrates authentication by delegating to registered Authenticators.
// It iterates through available authenticators to find one that supports the
// authentication kind and can successfully validate the credentials.
type AuthManager interface {
	// Authenticate finds a suitable authenticator and validates the credentials.
	Authenticate(ctx context.Context, authentication Authentication) (*Principal, error)
}

// UserLoader retrieves user information for authentication and authorization.
// Implementations typically query a database or external identity provider.
type UserLoader interface {
	// LoadByUsername retrieves a user by username, returning the Principal,
	// hashed password, and any error. Used for password-based authentication.
	LoadByUsername(ctx context.Context, username string) (*Principal, string, error)
	// LoadByID retrieves a user by their unique identifier.
	// Used for token refresh and session validation.
	LoadByID(ctx context.Context, id string) (*Principal, error)
}

// ExternalAppLoader retrieves external application credentials for API authentication.
// Used by OpenAPI authenticator to validate app-based signature authentication.
type ExternalAppLoader interface {
	// LoadByID retrieves an external app by its ID, returning the Principal,
	// secret key for signature verification, and any error.
	LoadByID(ctx context.Context, id string) (*Principal, string, error)
}

// PasswordDecryptor decrypts client-side encrypted passwords before verification.
// Used when passwords are encrypted during transmission for additional security.
type PasswordDecryptor interface {
	// Decrypt transforms an encrypted password back to plaintext for verification.
	Decrypt(encryptedPassword string) (string, error)
}

// PermissionChecker verifies if a Principal has a specific permission.
// Used by authorization middleware to enforce access control on API endpoints.
type PermissionChecker interface {
	// HasPermission returns true if the Principal has the specified permission token.
	HasPermission(ctx context.Context, principal *Principal, permToken string) (bool, error)
}

// RolePermissionsLoader retrieves permissions associated with a role.
// Used by RBAC implementations to build the permission set for authorization checks.
type RolePermissionsLoader interface {
	// LoadPermissions returns a map of permission tokens to their DataScope for a role.
	LoadPermissions(ctx context.Context, role string) (map[string]DataScope, error)
}

// UserInfoLoader retrieves extended user information for the current session.
// Used to populate user profile data, preferences, or other session-specific details.
type UserInfoLoader interface {
	// LoadUserInfo retrieves detailed user information based on the Principal and parameters.
	LoadUserInfo(ctx context.Context, principal *Principal, params map[string]any) (*UserInfo, error)
}

// DataScope defines row-level data access restrictions.
// Implementations filter query results based on the Principal's permissions,
// enabling multi-tenant data isolation or hierarchical data access control.
type DataScope interface {
	// Key returns a unique identifier for this data scope type.
	Key() string
	// Priority determines the order when multiple scopes apply (lower = higher priority).
	Priority() int
	// Supports returns true if this scope applies to the given Principal and table.
	Supports(principal *Principal, table *orm.Table) bool
	// Apply modifies the query to enforce the data scope restrictions.
	Apply(principal *Principal, query orm.SelectQuery) error
}

// DataPermissionResolver determines the applicable DataScope for a permission.
// Used to translate permission tokens into concrete data filtering rules.
type DataPermissionResolver interface {
	// ResolveDataScope returns the DataScope that should be applied for the permission.
	ResolveDataScope(ctx context.Context, principal *Principal, permToken string) (DataScope, error)
}

// DataPermissionApplier applies data permission filters to database queries.
// Wraps the resolution and application of DataScope into a single operation.
type DataPermissionApplier interface {
	// Apply adds data permission filters to the query based on the current context.
	Apply(query orm.SelectQuery) error
}

// NonceStore manages nonce lifecycle for replay attack prevention.
// Stores used nonces with TTL to detect and reject duplicate requests.
// Implementations must be thread-safe for concurrent access.
type NonceStore interface {
	// Exists checks if a nonce has already been used for the given app.
	Exists(ctx context.Context, appID, nonce string) (bool, error)
	// Store saves a nonce with the specified TTL.
	// The TTL should be slightly longer than timestamp tolerance to ensure
	// nonces remain valid while their corresponding timestamps are accepted.
	Store(ctx context.Context, appID, nonce string, ttl time.Duration) error
}

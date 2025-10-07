package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/set"
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
	// ctx: The request context for carrying trace info, cancellation signals, etc.
	// principal: The current user/app/system identity.
	// permissionToken: The permission token required by the API endpoint.
	// Returns true if the principal has the permission, false otherwise.
	// Returns an error if the permission check fails due to internal errors.
	HasPermission(ctx context.Context, principal *Principal, permissionToken string) (bool, error)
}

// RolePermissionsLoader defines a strategy for loading permissions associated with a role.
// This interface is used by the RBAC PermissionChecker implementation.
// Users should implement this interface to define how role permissions are loaded.
type RolePermissionsLoader interface {
	// LoadPermissions loads all permission tokens associated with the given role.
	// ctx: The request context.
	// role: The role name to load permissions for.
	// Returns a set of permission tokens for the role.
	// Returns an empty set if the role doesn't exist or has no permissions.
	// Returns an error if loading fails due to internal errors.
	LoadPermissions(ctx context.Context, role string) (set.Set[string], error)
}

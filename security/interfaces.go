package security

// TokenCredentials represents the authentication tokens for a user.
// It contains both access token and refresh token for token-based authentication.
type TokenCredentials struct {
	AccessToken  string `json:"accessToken"`  // AccessToken is the short-lived access token for API requests
	RefreshToken string `json:"refreshToken"` // RefreshToken is the long-lived token used to refresh access tokens
}

// Authentication represents the authentication information provided by a client.
// It contains the authentication type, principal identifier, and credentials.
type Authentication struct {
	Type        string `json:"type"`        // Type specifies the authentication method (e.g., "password", "jwt", "oauth")
	Principal   string `json:"principal"`   // Principal is the identifier of the entity being authenticated (e.g., username, email)
	Credentials any    `json:"credentials"` // Credentials contains the authentication data (e.g., password, token)
}

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
	Generate(principal *Principal) (*TokenCredentials, error)
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

// ExternalAppConfig is an optional details payload for openapi principals.
// Users can place this config into Principal.Details for extra runtime checks.
type ExternalAppConfig struct {
	Enabled     bool   `json:"enabled"`     // Enabled indicates whether the external app is enabled
	IpWhitelist string `json:"ipWhitelist"` // IpWhitelist is a comma-separated whitelist (supports IP or CIDR)
}

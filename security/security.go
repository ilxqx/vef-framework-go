package security

// AuthTokens represents the authentication tokens for a user.
// It contains both access token and refresh token for token-based authentication.
type AuthTokens struct {
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

// ExternalAppConfig is an optional details payload for openapi principals.
// Users can place this config into Principal.Details for extra runtime checks.
type ExternalAppConfig struct {
	Enabled     bool   `json:"enabled"`     // Enabled indicates whether the external app is enabled
	IpWhitelist string `json:"ipWhitelist"` // IpWhitelist is a comma-separated whitelist (supports IP or CIDR)
}

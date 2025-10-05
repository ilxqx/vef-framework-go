package security

import (
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// AuthenticatorAuthManager implements the AuthManager interface.
// It manages multiple authenticators and delegates authentication requests to the appropriate one.
type AuthenticatorAuthManager struct {
	authenticators []security.Authenticator
}

// NewAuthManager creates a new authentication manager with the provided authenticators.
func NewAuthManager(authenticators []security.Authenticator) security.AuthManager {
	return &AuthenticatorAuthManager{
		authenticators: authenticators,
	}
}

// Authenticate attempts to authenticate the provided authentication information.
// It finds the appropriate authenticator and delegates the authentication request.
func (am *AuthenticatorAuthManager) Authenticate(authentication security.Authentication) (*security.Principal, error) {
	// Find the appropriate authenticator
	authenticator := am.findAuthenticator(authentication.Type)
	if authenticator == nil {
		logger.Warnf("No authenticator found for authentication type: %s", authentication.Type)

		return nil, result.ErrWithCodef(
			result.ErrCodeUnsupportedAuthenticationType,
			"Authentication type '%s' is not supported",
			authentication.Type,
		)
	}

	// Delegate to the authenticator
	principal, err := authenticator.Authenticate(authentication)
	if err != nil {
		logger.Warnf("Authentication failed for principal '%s' with type '%s': %v",
			authentication.Principal, authentication.Type, err)

		return nil, err
	}

	logger.Infof("Authentication successful for principal '%s' with type '%s'",
		authentication.Principal, authentication.Type)

	return principal, nil
}

// findAuthenticator finds the first authenticator that supports the given authentication type.
func (am *AuthenticatorAuthManager) findAuthenticator(authType string) security.Authenticator {
	for _, authenticator := range am.authenticators {
		if authenticator.Supports(authType) {
			return authenticator
		}
	}

	return nil
}

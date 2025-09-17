package security

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// defaultAuthManager implements the AuthManager interface.
// It manages multiple authenticators and delegates authentication requests to the appropriate one.
type defaultAuthManager struct {
	authenticators []security.Authenticator
}

// newAuthManager creates a new authentication manager with the provided authenticators.
func newAuthManager(authenticators []security.Authenticator) security.AuthManager {
	return &defaultAuthManager{
		authenticators: authenticators,
	}
}

// Authenticate attempts to authenticate the provided authentication information.
// It finds the appropriate authenticator and delegates the authentication request.
func (am *defaultAuthManager) Authenticate(authentication security.Authentication) (*security.Principal, error) {
	// Find the appropriate authenticator
	authenticator := am.findAuthenticator(authentication.Type)
	if authenticator == nil {
		logger.Warnf("No authenticator found for authentication type: %s", authentication.Type)
		return nil, result.ErrWithCode(
			result.ErrCodeUnsupportedAuthenticationType,
			fmt.Sprintf("Authentication type '%s' is not supported", authentication.Type),
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
func (am *defaultAuthManager) findAuthenticator(authType string) security.Authenticator {
	for _, authenticator := range am.authenticators {
		if authenticator.Supports(authType) {
			return authenticator
		}
	}

	return nil
}

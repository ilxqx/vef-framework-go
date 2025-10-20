package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/constants"
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
func (am *AuthenticatorAuthManager) Authenticate(ctx context.Context, authentication security.Authentication) (*security.Principal, error) {
	// Find the appropriate authenticator
	authenticator := am.findAuthenticator(authentication.Type)
	if authenticator == nil {
		logger.Warnf("No authenticator found for authentication type: %s", authentication.Type)

		return nil, result.ErrWithCodef(
			result.ErrCodeUnsupportedAuthenticationType,
			"Authentication type %q is not supported",
			authentication.Type,
		)
	}

	// Delegate to the authenticator
	principal, err := authenticator.Authenticate(ctx, authentication)
	if err != nil {
		if _, ok := result.AsErr(err); !ok {
			// Mask sensitive principal information for security
			maskedPrincipal := maskPrincipal(authentication.Principal)
			logger.Warnf("Authentication failed: type=%s, principal=%s, authenticator=%T, error=%v",
				authentication.Type, maskedPrincipal, authenticator, err)
		}

		return nil, err
	}

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

// maskPrincipal masks sensitive principal information for logging.
// It shows the first 3 characters and masks the rest with asterisks.
func maskPrincipal(principal string) string {
	if principal == constants.Empty {
		return "<empty>"
	}

	length := len(principal)
	if length <= 3 {
		return "***"
	}

	// Show first 3 characters, mask the rest
	return principal[:3] + "***"
}

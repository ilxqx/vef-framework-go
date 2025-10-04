package security

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	// Token authentication type
	AuthTypeToken = "token"
)

// JWTTokenAuthenticator implements the Authenticator interface for JWT token authentication.
// It validates JWT tokens and extracts principal information from them.
type JWTTokenAuthenticator struct {
	jwt *security.JWT
}

// NewJWTAuthenticator creates a new JWT authenticator.
func NewJWTAuthenticator(jwt *security.JWT) security.Authenticator {
	return &JWTTokenAuthenticator{
		jwt: jwt,
	}
}

// Supports checks if this authenticator can handle JWT authentication.
func (*JWTTokenAuthenticator) Supports(authType string) bool {
	return authType == AuthTypeToken
}

// Authenticate validates the JWT token and returns the principal.
// The credentials field should contain the JWT access token.
func (ja *JWTTokenAuthenticator) Authenticate(authentication security.Authentication) (*security.Principal, error) {
	// Extract the token from credentials
	token := authentication.Principal
	if token == constants.Empty {
		return nil, result.ErrWithCode(
			result.ErrCodePrincipalInvalid,
			i18n.T("token_invalid"),
		)
	}

	// Parse the JWT access token
	claimsAccessor, err := ja.jwt.Parse(token)
	if err != nil {
		logger.Warnf("JWT token validation failed: %v", err)
		return nil, err
	}

	if claimsAccessor.Type() != tokenTypeAccess {
		return nil, result.ErrWithCode(
			result.ErrCodeTokenInvalid,
			i18n.T("token_invalid"),
		)
	}

	// Subject format: id@name, where '@' is defined by constants.At
	subjectParts := strings.SplitN(claimsAccessor.Subject(), constants.At, 2)
	principal := security.NewUser(subjectParts[0], subjectParts[1], claimsAccessor.Roles()...)
	principal.AttemptUnmarshalDetails(claimsAccessor.Details())

	logger.Infof("JWT authentication successful for principal '%s'", principal.Id)
	return principal, nil
}

package security

import (
	"context"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	// Token authentication type.
	AuthTypeToken = "token"
)

// JwtTokenAuthenticator implements the Authenticator interface for Jwt token authentication.
// It validates Jwt tokens and extracts principal information from them.
type JwtTokenAuthenticator struct {
	jwt *security.Jwt
}

// NewJwtAuthenticator creates a new Jwt authenticator.
func NewJwtAuthenticator(jwt *security.Jwt) security.Authenticator {
	return &JwtTokenAuthenticator{
		jwt: jwt,
	}
}

// Supports checks if this authenticator can handle Jwt authentication.
func (*JwtTokenAuthenticator) Supports(authType string) bool {
	return authType == AuthTypeToken
}

// Authenticate validates the Jwt token and returns the principal.
// The credentials field should contain the Jwt access token.
func (ja *JwtTokenAuthenticator) Authenticate(_ context.Context, authentication security.Authentication) (*security.Principal, error) {
	// Extract the token from credentials
	token := authentication.Principal
	if token == constants.Empty {
		return nil, result.Err(
			i18n.T("token_invalid"),
			result.WithCode(result.ErrCodePrincipalInvalid),
		)
	}

	// Parse the Jwt access token
	claimsAccessor, err := ja.jwt.Parse(token)
	if err != nil {
		return nil, err
	}

	if claimsAccessor.Type() != tokenTypeAccess {
		return nil, result.Err(
			i18n.T("token_invalid"),
			result.WithCode(result.ErrCodeTokenInvalid),
		)
	}

	// Subject format: id@name, where '@' is defined by constants.At
	subjectParts := strings.SplitN(claimsAccessor.Subject(), constants.At, 2)
	principal := security.NewUser(subjectParts[0], subjectParts[1], claimsAccessor.Roles()...)
	principal.AttemptUnmarshalDetails(claimsAccessor.Details())

	return principal, nil
}

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
	AuthTypeToken = "token"
)

type JwtTokenAuthenticator struct {
	jwt *security.Jwt
}

func NewJwtAuthenticator(jwt *security.Jwt) security.Authenticator {
	return &JwtTokenAuthenticator{
		jwt: jwt,
	}
}

func (*JwtTokenAuthenticator) Supports(authType string) bool {
	return authType == AuthTypeToken
}

func (ja *JwtTokenAuthenticator) Authenticate(_ context.Context, authentication security.Authentication) (*security.Principal, error) {
	token := authentication.Principal
	if token == constants.Empty {
		return nil, result.Err(
			i18n.T("token_invalid"),
			result.WithCode(result.ErrCodePrincipalInvalid),
		)
	}

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

	subjectParts := strings.SplitN(claimsAccessor.Subject(), constants.At, 2)
	principal := security.NewUser(subjectParts[0], subjectParts[1], claimsAccessor.Roles()...)
	principal.AttemptUnmarshalDetails(claimsAccessor.Details())

	return principal, nil
}

package security

import (
	"context"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	AuthKindToken = "token"
)

type JWTTokenAuthenticator struct {
	jwt *security.JWT
}

func NewJWTAuthenticator(jwt *security.JWT) security.Authenticator {
	return &JWTTokenAuthenticator{
		jwt: jwt,
	}
}

func (*JWTTokenAuthenticator) Supports(kind string) bool {
	return kind == AuthKindToken
}

func (ja *JWTTokenAuthenticator) Authenticate(_ context.Context, authentication security.Authentication) (*security.Principal, error) {
	token := authentication.Principal
	if token == constants.Empty {
		return nil, result.ErrTokenInvalid
	}

	claimsAccessor, err := ja.jwt.Parse(token)
	if err != nil {
		return nil, err
	}

	if claimsAccessor.Type() != tokenTypeAccess {
		return nil, result.ErrTokenInvalid
	}

	subjectParts := strings.SplitN(claimsAccessor.Subject(), constants.At, 2)
	principal := security.NewUser(subjectParts[0], subjectParts[1], claimsAccessor.Roles()...)
	principal.AttemptUnmarshalDetails(claimsAccessor.Details())

	return principal, nil
}

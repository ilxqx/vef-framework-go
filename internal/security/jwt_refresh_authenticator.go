package security

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	// JWT refresh authentication type
	AuthTypeJWTRefresh = "jwt_refresh"
)

type JWTRefreshAuthenticator struct {
	jwt *security.JWT
}

func NewJWTRefreshAuthenticator(jwt *security.JWT) security.Authenticator {
	return &JWTRefreshAuthenticator{
		jwt: jwt,
	}
}

func (j *JWTRefreshAuthenticator) Supports(authType string) bool {
	return authType == AuthTypeJWTRefresh
}

func (j *JWTRefreshAuthenticator) Authenticate(authentication security.Authentication) (*security.Principal, error) {
	token := authentication.Principal
	if token == constants.Empty {
		return nil, result.ErrWithCode(
			result.ErrCodePrincipalInvalid,
			"令牌不能为空",
		)
	}

	// Parse the JWT refresh token
	claimsAccessor, err := j.jwt.Parse(token)
	if err != nil {
		logger.Warnf("JWT refresh token validation failed: %v", err)
		return nil, err
	}

	if claimsAccessor.Type() != tokenTypeRefresh {
		return nil, result.ErrWithCode(
			result.ErrCodeTokenInvalid,
			"非法令牌类型",
		)
	}

	// Subject format: id@name, where '@' is defined by constants.At
	subjectParts := strings.SplitN(claimsAccessor.Subject(), constants.At, 2)
	principal := security.NewUser(subjectParts[0], subjectParts[1])
	return principal, nil
}

package security

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	// Refresh authentication type
	AuthTypeRefresh = "refresh"
)

type JWTRefreshAuthenticator struct {
	jwt        *security.JWT
	userLoader security.UserLoader
}

func NewJWTRefreshAuthenticator(jwt *security.JWT, userLoader security.UserLoader) security.Authenticator {
	return &JWTRefreshAuthenticator{
		jwt:        jwt,
		userLoader: userLoader,
	}
}

func (j *JWTRefreshAuthenticator) Supports(authType string) bool {
	return authType == AuthTypeRefresh
}

func (j *JWTRefreshAuthenticator) Authenticate(authentication security.Authentication) (*security.Principal, error) {
	if j.userLoader == nil {
		return nil, result.ErrWithCode(result.ErrCodeNotImplemented, i18n.T("user_loader_not_implemented"))
	}

	token := authentication.Principal
	if token == constants.Empty {
		return nil, result.ErrWithCode(
			result.ErrCodePrincipalInvalid,
			i18n.T("token_invalid"),
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
			i18n.T("token_invalid"),
		)
	}

	// Subject format: id@name, where '@' is defined by constants.At
	subjectParts := strings.SplitN(claimsAccessor.Subject(), constants.At, 2)
	userId := subjectParts[0]

	// Reload the latest user data by ID to ensure current user state (permissions, status, etc.)
	principal, err := j.userLoader.LoadById(userId)
	if err != nil {
		logger.Warnf("Failed to reload user by Id '%s': %v", userId, err)
		return nil, err
	}
	if principal == nil {
		logger.Warnf("User not found by Id '%s'", userId)
		return nil, result.ErrWithCode(result.ErrCodeRecordNotFound, i18n.T("record_not_found"))
	}

	return principal, nil
}

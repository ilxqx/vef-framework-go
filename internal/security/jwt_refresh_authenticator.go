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
	AuthKindRefresh = "refresh"
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

func (j *JWTRefreshAuthenticator) Supports(kind string) bool {
	return kind == AuthKindRefresh
}

func (j *JWTRefreshAuthenticator) Authenticate(ctx context.Context, authentication security.Authentication) (*security.Principal, error) {
	if j.userLoader == nil {
		return nil, result.ErrNotImplemented(i18n.T(result.ErrMessageUserLoaderNotImplemented))
	}

	token := authentication.Principal
	if token == constants.Empty {
		return nil, result.ErrTokenInvalid
	}

	claimsAccessor, err := j.jwt.Parse(token)
	if err != nil {
		logger.Warnf("JWT refresh token validation failed: %v", err)

		return nil, err
	}

	if claimsAccessor.Type() != tokenTypeRefresh {
		return nil, result.ErrTokenInvalid
	}

	subjectParts := strings.SplitN(claimsAccessor.Subject(), constants.At, 2)
	userID := subjectParts[0]

	// Reload user to get latest permissions/status instead of relying on stale token data.
	principal, err := j.userLoader.LoadByID(ctx, userID)
	if err != nil {
		logger.Warnf("Failed to reload user by ID %q: %v", userID, err)

		return nil, err
	}

	if principal == nil {
		logger.Warnf("User not found by ID %q", userID)

		return nil, result.ErrRecordNotFound
	}

	return principal, nil
}

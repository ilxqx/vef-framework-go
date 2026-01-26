package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/password"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	AuthKindPassword = "password"
)

// PasswordAuthenticator verifies username/password credentials with optional decryption support
// for scenarios where clients encrypt passwords before transmission.
type PasswordAuthenticator struct {
	loader  security.UserLoader
	encoder password.Encoder
}

func NewPasswordAuthenticator(
	loader security.UserLoader,
	encoder password.Encoder,
) security.Authenticator {
	return &PasswordAuthenticator{
		loader:  loader,
		encoder: encoder,
	}
}

func (*PasswordAuthenticator) Supports(kind string) bool { return kind == AuthKindPassword }

func (p *PasswordAuthenticator) Authenticate(ctx context.Context, authentication security.Authentication) (*security.Principal, error) {
	if p.loader == nil {
		return nil, result.ErrNotImplemented(i18n.T(result.ErrMessageUserLoaderNotImplemented))
	}

	username := authentication.Principal
	if username == constants.Empty {
		return nil, result.ErrPrincipalInvalid(i18n.T("username_required"))
	}

	if username == constants.PrincipalSystem || username == constants.PrincipalCronJob || username == constants.PrincipalAnonymous {
		return nil, result.ErrPrincipalInvalid(i18n.T("system_principal_login_forbidden"))
	}

	password, ok := authentication.Credentials.(string)
	if !ok || password == constants.Empty {
		return nil, result.ErrCredentialsInvalid(i18n.T("password_required"))
	}

	principal, passwordHash, err := p.loader.LoadByUsername(ctx, username)
	if err != nil {
		if result.IsRecordNotFound(err) {
			logger.Infof("User loader returned record not found for username %q", username)
		} else {
			logger.Warnf("Failed to load user by username %q: %v", username, err)
		}

		return nil, result.ErrCredentialsInvalid(i18n.T("invalid_credentials"))
	}

	if principal == nil || passwordHash == constants.Empty {
		return nil, result.ErrCredentialsInvalid(i18n.T("invalid_credentials"))
	}

	if !p.encoder.Matches(password, passwordHash) {
		return nil, result.ErrCredentialsInvalid(i18n.T("invalid_credentials"))
	}

	logger.Infof("Password authentication successful for principal %q", principal.ID)

	return principal, nil
}

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
	AuthTypePassword = "password"
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

func (*PasswordAuthenticator) Supports(authType string) bool { return authType == AuthTypePassword }

func (p *PasswordAuthenticator) Authenticate(ctx context.Context, authentication security.Authentication) (*security.Principal, error) {
	if p.loader == nil {
		return nil, result.Err(i18n.T("user_loader_not_implemented"), result.WithCode(result.ErrCodeNotImplemented))
	}

	username := authentication.Principal
	if username == constants.Empty {
		return nil, result.Err(i18n.T("username_required"), result.WithCode(result.ErrCodePrincipalInvalid))
	}

	if username == constants.PrincipalSystem || username == constants.PrincipalCronJob || username == constants.PrincipalAnonymous {
		return nil, result.Err(i18n.T("system_principal_login_forbidden"), result.WithCode(result.ErrCodePrincipalInvalid))
	}

	password, ok := authentication.Credentials.(string)
	if !ok || password == constants.Empty {
		return nil, result.Err(i18n.T("password_required"), result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	principal, passwordHash, err := p.loader.LoadByUsername(ctx, username)
	if err != nil {
		if result.IsRecordNotFound(err) {
			logger.Infof("User loader returned record not found for username %q", username)
		} else {
			logger.Warnf("Failed to load user by username %q: %v", username, err)
		}

		return nil, result.Err(i18n.T("invalid_credentials"), result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	if principal == nil || passwordHash == constants.Empty {
		return nil, result.Err(i18n.T("invalid_credentials"), result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	if !p.encoder.Matches(password, passwordHash) {
		return nil, result.Err(i18n.T("invalid_credentials"), result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	logger.Infof("Password authentication successful for principal %q", principal.Id)

	return principal, nil
}

package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	// Password authentication type.
	AuthTypePassword = "password"
)

// PasswordAuthenticator implements the Authenticator interface using username/password verification.
// It relies on an externally provided security.UserLoader to load user info and password hash.
// Optionally supports password decryption via security.PasswordDecryptor for scenarios where
// clients encrypt passwords before transmission.
type PasswordAuthenticator struct {
	loader    security.UserLoader
	decryptor security.PasswordDecryptor
}

// NewPasswordAuthenticator creates a new password authenticator with the given user loader.
// The decryptor parameter is optional; pass nil if passwords are transmitted in plaintext.
func NewPasswordAuthenticator(loader security.UserLoader, decryptor security.PasswordDecryptor) security.Authenticator {
	return &PasswordAuthenticator{
		loader:    loader,
		decryptor: decryptor,
	}
}

// Supports checks if this authenticator can handle password authentication.
func (*PasswordAuthenticator) Supports(authType string) bool { return authType == AuthTypePassword }

// Authenticate validates credentials which should be a plaintext password for the given principal (username).
func (p *PasswordAuthenticator) Authenticate(ctx context.Context, authentication security.Authentication) (*security.Principal, error) {
	if p.loader == nil {
		return nil, result.Err(i18n.T("user_loader_not_implemented"), result.WithCode(result.ErrCodeNotImplemented))
	}

	username := authentication.Principal
	if username == constants.Empty {
		return nil, result.Err(i18n.T("username_required"), result.WithCode(result.ErrCodePrincipalInvalid))
	}

	// Prevent system internal principals from logging in
	if username == constants.PrincipalSystem || username == constants.PrincipalCronJob || username == constants.PrincipalAnonymous {
		return nil, result.Err(i18n.T("system_principal_login_forbidden"), result.WithCode(result.ErrCodePrincipalInvalid))
	}

	// Expect password in credentials (may be encrypted)
	password, ok := authentication.Credentials.(string)
	if !ok || password == constants.Empty {
		return nil, result.Err(i18n.T("password_required"), result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	// Decrypt password if decryptor is provided
	if p.decryptor != nil {
		plaintextPassword, err := p.decryptor.Decrypt(password)
		if err != nil {
			logger.Errorf("Failed to decrypt password for principal %q: %v", username, err)

			return nil, result.Err(i18n.T("invalid_credentials"), result.WithCode(result.ErrCodeCredentialsInvalid))
		}

		password = plaintextPassword
	}

	// Load user info and password hash via injected loader
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

	// Compare password with stored hash
	if !security.VerifyPassword(password, passwordHash) {
		return nil, result.Err(i18n.T("invalid_credentials"), result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	logger.Infof("Password authentication successful for principal %q", principal.Id)

	return principal, nil
}

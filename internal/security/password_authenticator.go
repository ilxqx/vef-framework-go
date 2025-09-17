package security

import (
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/utils"
)

const (
	AuthTypePassword = "password" // Password authentication type
)

// passwordAuthenticator implements the Authenticator interface using username/password verification.
// It relies on an externally provided security.UserLoader to load user info and password hash.
type passwordAuthenticator struct {
	loader security.UserLoader // loader loads user principal and hashed password by username
}

// newPasswordAuthenticator creates a new password authenticator with the given user loader.
func newPasswordAuthenticator(loader security.UserLoader) security.Authenticator {
	return &passwordAuthenticator{loader: loader}
}

// Supports checks if this authenticator can handle password authentication.
func (*passwordAuthenticator) Supports(authType string) bool { return authType == AuthTypePassword }

// Authenticate validates credentials which should be a plaintext password for the given principal (username).
func (a *passwordAuthenticator) Authenticate(authentication security.Authentication) (*security.Principal, error) {
	if a.loader == nil {
		return nil, result.ErrWithCode(result.ErrCodeNotImplemented, "请提供一个用户加载器实现")
	}
	username := authentication.Principal
	if username == constants.Empty {
		return nil, result.ErrWithCode(result.ErrCodePrincipalInvalid, "账号不能为空")
	}

	// Expect plaintext password in credentials
	password, ok := authentication.Credentials.(string)
	if !ok || password == constants.Empty {
		return nil, result.ErrWithCode(result.ErrCodeCredentialsInvalid, "密码不能为空")
	}

	// Load user info and password hash via injected loader
	principal, passwordHash, err := a.loader.LoadByUsername(username)
	if err != nil {
		return nil, err
	}
	if principal == nil || passwordHash == constants.Empty {
		return nil, result.ErrWithCode(result.ErrCodeCredentialsInvalid, "账号或密码错误")
	}

	// Compare password with stored hash
	if utils.VerifyPassword(password, passwordHash) {
		return nil, result.ErrWithCode(result.ErrCodeCredentialsInvalid, "账号或密码错误")
	}

	logger.Infof("password authentication successful for principal '%s'", principal.Id)
	return principal, nil
}

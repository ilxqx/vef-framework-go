package security

import (
	"context"
	"errors"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// AuthKindSignature is the authentication kind for signature-based authentication.
const AuthKindSignature = "signature"

// SignatureAuthenticator validates HMAC-based signatures for external app authentication.
type SignatureAuthenticator struct {
	loader  security.ExternalAppLoader
	options []security.SignatureOption
}

// NewSignatureAuthenticator creates a new signature authenticator.
// Returns nil if loader is nil (optional dependency pattern for FX).
func NewSignatureAuthenticator(
	loader security.ExternalAppLoader,
	nonceStore security.NonceStore,
) (security.Authenticator, error) {
	if loader == nil {
		return nil, nil
	}

	var options []security.SignatureOption
	if nonceStore != nil {
		options = append(options, security.WithNonceStore(nonceStore))
	}

	return &SignatureAuthenticator{
		loader:  loader,
		options: options,
	}, nil
}

func (*SignatureAuthenticator) Supports(kind string) bool {
	return kind == AuthKindSignature
}

func (a *SignatureAuthenticator) Authenticate(ctx context.Context, authentication security.Authentication) (*security.Principal, error) {
	appID := authentication.Principal
	if appID == constants.Empty {
		return nil, result.ErrAppIDRequired
	}

	credentials, ok := authentication.Credentials.(*security.SignatureCredentials)
	if !ok || credentials == nil {
		return nil, result.ErrCredentialsInvalid(i18n.T(result.ErrMessageCredentialsFormatInvalid))
	}

	principal, secret, err := a.loader.LoadByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	if principal == nil || secret == constants.Empty {
		return nil, result.ErrExternalAppNotFound
	}

	if err := a.validateIPWhitelist(ctx, principal); err != nil {
		return nil, err
	}

	if err := a.verifySignature(ctx, appID, secret, credentials); err != nil {
		return nil, err
	}

	logger.Infof("Signature authentication successful for app %q", principal.ID)

	return principal, nil
}

func (a *SignatureAuthenticator) verifySignature(
	ctx context.Context,
	appID, secret string,
	credentials *security.SignatureCredentials,
) error {
	sig, err := security.NewSignature(secret, a.options...)
	if err != nil {
		return mapSignatureError(err)
	}

	if err := sig.Verify(ctx, appID, credentials.Timestamp, credentials.Nonce, credentials.Signature); err != nil {
		return mapSignatureError(err)
	}

	return nil
}

func (*SignatureAuthenticator) validateIPWhitelist(ctx context.Context, principal *security.Principal) error {
	details, ok := principal.Details.(*security.ExternalAppConfig)
	if !ok || details == nil {
		return nil
	}

	if !details.Enabled {
		return result.ErrExternalAppDisabled
	}

	if details.IPWhitelist == constants.Empty {
		return nil
	}

	requestIP := contextx.RequestIP(ctx)
	if requestIP == constants.Empty {
		return nil
	}

	if validator := security.NewIPWhitelistValidator(details.IPWhitelist); !validator.IsAllowed(requestIP) {
		return result.ErrIPNotAllowed
	}

	return nil
}

// mapSignatureError converts security.Signature errors to result errors.
func mapSignatureError(err error) error {
	switch {
	case errors.Is(err, security.ErrSignatureExpired):
		return result.ErrSignatureExpired
	case errors.Is(err, security.ErrSignatureInvalid):
		return result.ErrSignatureInvalid
	case errors.Is(err, security.ErrSignatureNonceUsed):
		return result.ErrNonceAlreadyUsed
	case errors.Is(err, security.ErrSignatureNonceRequired):
		return result.ErrNonceRequired
	default:
		return result.ErrSignatureInvalid
	}
}

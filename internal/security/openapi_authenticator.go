package security

import (
	"context"
	"crypto/hmac"
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/hashx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	AuthKindOpenApi = "openapi"
)

// OpenApiAuthenticator validates HMAC-based signatures for external app authentication.
// Credentials format: "<signatureHex>@<timestamp>@<bodySha256Base64>".
// Signature: hex(HMAC-SHA256(secret, appID + "\n" + timestamp + "\n" + bodySha256Base64)).
type OpenApiAuthenticator struct {
	loader security.ExternalAppLoader
}

func NewOpenApiAuthenticator(loader security.ExternalAppLoader) security.Authenticator {
	return &OpenApiAuthenticator{loader: loader}
}

func (*OpenApiAuthenticator) Supports(kind string) bool { return kind == AuthKindOpenApi }

func (a *OpenApiAuthenticator) Authenticate(ctx context.Context, authentication security.Authentication) (*security.Principal, error) {
	if a.loader == nil {
		return nil, result.ErrNotImplemented(i18n.T(result.ErrMessageExternalAppLoaderNotImplemented))
	}

	appID := authentication.Principal
	if appID == constants.Empty {
		return nil, result.ErrAppIDRequired
	}

	cred, ok := authentication.Credentials.(string)
	if !ok || cred == constants.Empty {
		return nil, result.ErrSignatureRequired
	}

	parts := strings.SplitN(cred, constants.At, 3)
	if len(parts) != 3 {
		return nil, result.ErrCredentialsInvalid(i18n.T(result.ErrMessageCredentialsFormatInvalid))
	}

	signatureHex, timestamp, bodyHash := parts[0], parts[1], parts[2]
	if signatureHex == constants.Empty || timestamp == constants.Empty || bodyHash == constants.Empty {
		return nil, result.ErrCredentialsInvalid(i18n.T(result.ErrMessageCredentialsFieldsRequired))
	}

	principal, secret, err := a.loader.LoadByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	if principal == nil || secret == constants.Empty {
		return nil, result.ErrExternalAppNotFound
	}

	signaturePayload := fmt.Sprintf("%s\n%s\n%s", appID, timestamp, bodyHash)

	secretBytes, err := encoding.FromHex(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode app secret: %w", err)
	}

	expectedSignatureHex := hashx.HmacSHA256(secretBytes, []byte(signaturePayload))

	providedMac, err := encoding.FromHex(signatureHex)
	if err != nil {
		return nil, result.ErrCredentialsInvalid(i18n.T(result.ErrMessageSignatureDecodeFailed))
	}

	expectedMac, err := encoding.FromHex(expectedSignatureHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode expected signature: %w", err)
	}

	if !hmac.Equal(expectedMac, providedMac) {
		return nil, result.ErrSignatureInvalid
	}

	logger.Infof("Openapi authentication successful for principal %q", principal.ID)

	return principal, nil
}

package security

import (
	"context"
	"crypto/hmac"
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/hash"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	AuthTypeOpenApi = "openapi"
)

// OpenApiAuthenticator validates HMAC-based signatures for external app authentication.
// Credentials format: "<signatureHex>@<timestamp>@<bodySha256Base64>".
// Signature: hex(HMAC-SHA256(secret, appId + "\n" + timestamp + "\n" + bodySha256Base64)).
type OpenApiAuthenticator struct {
	loader security.ExternalAppLoader
}

func NewOpenApiAuthenticator(loader security.ExternalAppLoader) security.Authenticator {
	return &OpenApiAuthenticator{loader: loader}
}

func (*OpenApiAuthenticator) Supports(authType string) bool { return authType == AuthTypeOpenApi }

func (a *OpenApiAuthenticator) Authenticate(ctx context.Context, authentication security.Authentication) (*security.Principal, error) {
	if a.loader == nil {
		return nil, result.Err(i18n.T("external_app_loader_not_implemented"), result.WithCode(result.ErrCodeNotImplemented))
	}

	appId := authentication.Principal
	if appId == constants.Empty {
		return nil, result.Err(result.ErrMessageAppIdRequired, result.WithCode(result.ErrCodeAppIdRequired))
	}

	cred, ok := authentication.Credentials.(string)
	if !ok || cred == constants.Empty {
		return nil, result.Err(result.ErrMessageSignatureRequired, result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	parts := strings.SplitN(cred, constants.At, 3)
	if len(parts) != 3 {
		return nil, result.Err(i18n.T("credentials_format_invalid"), result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	signatureHex := parts[0]
	timestamp := parts[1]

	bodyHash := parts[2]
	if signatureHex == constants.Empty || timestamp == constants.Empty || bodyHash == constants.Empty {
		return nil, result.Err(i18n.T("credentials_fields_required"), result.WithCode(result.ErrCodeCredentialsInvalid))
	}

	principal, secret, err := a.loader.LoadById(ctx, appId)
	if err != nil {
		return nil, err
	}

	if principal == nil || secret == constants.Empty {
		return nil, result.Err(result.ErrMessageExternalAppNotFound, result.WithCode(result.ErrCodeExternalAppNotFound))
	}

	var sb strings.Builder

	_, _ = sb.WriteString(appId)
	_ = sb.WriteByte(constants.ByteNewline)
	_, _ = sb.WriteString(timestamp)
	_ = sb.WriteByte(constants.ByteNewline)
	_, _ = sb.WriteString(bodyHash)

	secretBytes, err := encoding.FromHex(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode app secret: %w", err)
	}

	expectedSignatureHex, err := hash.Sha256Hmac(secretBytes, []byte(sb.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to compute signature: %w", err)
	}

	providedMac, err := encoding.FromHex(signatureHex)
	if err != nil {
		return nil, result.Err(i18n.T("signature_decode_failed"), result.WithCode(result.ErrCodeSignatureInvalid))
	}

	expectedMac, err := encoding.FromHex(expectedSignatureHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode expected signature: %w", err)
	}

	if !hmac.Equal(expectedMac, providedMac) {
		return nil, result.ErrSignatureInvalid
	}

	logger.Infof("Openapi authentication successful for principal %q", principal.Id)

	return principal, nil
}

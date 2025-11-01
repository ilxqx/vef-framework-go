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
	// OpenApi authentication type.
	AuthTypeOpenApi = "openapi"
)

// OpenApiAuthenticator implements Authenticator for simple HMAC based OpenApi authentication.
// Contract in this framework:
//   - Authentication.Principal: appId
//   - Authentication.Credentials: "<signatureHex>@<timestamp>@<bodySha256Base64>"
//   - SignatureHex is computed as hex(HMAC-SHA256(secret, appId + "\n" + timestamp + "\n" + bodySha256Base64))
//     where bodySha256Base64 is the Base64(SHA256(raw request body)), timestamp is unix seconds string.
//   - We only consider appId, timestamp and body hash because the framework uses unified POST with Request body.
type OpenApiAuthenticator struct {
	loader security.ExternalAppLoader // loader loads external app principal and secret by appId
}

// NewOpenApiAuthenticator creates a new OpenApi authenticator with the given loader.
func NewOpenApiAuthenticator(loader security.ExternalAppLoader) security.Authenticator {
	return &OpenApiAuthenticator{loader: loader}
}

// Supports checks if this authenticator can handle OpenApi authentication.
func (*OpenApiAuthenticator) Supports(authType string) bool { return authType == AuthTypeOpenApi }

// Authenticate validates the provided OpenApi authentication information.
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

	// credentials format: "<signatureHex>@<timestamp>@<bodySha256Base64>"
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

	// Recompute signature: hex(HMAC-SHA256(secret, appId + "\n" + timestamp + "\n" + bodyHash))
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

	// Use hash package's HMAC-SHA256 function
	expectedSignatureHex := hash.Sha256Hmac(secretBytes, []byte(sb.String()))

	// Compare signatures using constant-time comparison
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

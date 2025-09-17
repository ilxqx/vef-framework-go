package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	AuthTypeOpenAPI = "openapi" // OpenAPI authentication type
)

// openapiAuthenticator implements Authenticator for simple HMAC based OpenAPI authentication.
// Contract in this framework:
//   - Authentication.Principal: appId
//   - Authentication.Credentials: "<signatureHex>@<timestamp>@<bodySha256Base64>"
//   - SignatureHex is computed as hex(HMAC-SHA256(secret, appId + "\n" + timestamp + "\n" + bodySha256Base64))
//     where bodySha256Base64 is the Base64(SHA256(raw request body)), timestamp is unix seconds string.
//   - We only consider appId, timestamp and body hash because the framework uses unified POST with Request body.
type openapiAuthenticator struct {
	loader security.ExternalAppLoader // loader loads external app principal and secret by appId
}

// newOpenAPIAuthenticator creates a new OpenAPI authenticator with the given loader.
func newOpenAPIAuthenticator(loader security.ExternalAppLoader) security.Authenticator {
	return &openapiAuthenticator{loader: loader}
}

// Supports checks if this authenticator can handle OpenAPI authentication.
func (*openapiAuthenticator) Supports(authType string) bool { return authType == AuthTypeOpenAPI }

// Authenticate validates the provided OpenAPI authentication information.
func (a *openapiAuthenticator) Authenticate(authentication security.Authentication) (*security.Principal, error) {
	if a.loader == nil {
		return nil, result.ErrWithCode(result.ErrCodeNotImplemented, "请提供一个 ExternalAppLoader 的实现")
	}

	appId := authentication.Principal
	if appId == constants.Empty {
		return nil, result.ErrWithCode(result.ErrCodeAppIdRequired, result.ErrMessageAppIdRequired)
	}

	cred, ok := authentication.Credentials.(string)
	if !ok || cred == constants.Empty {
		return nil, result.ErrWithCode(result.ErrCodeCredentialsInvalid, result.ErrMessageSignatureRequired)
	}

	// credentials format: "<signatureHex>@<timestamp>@<bodySha256Base64>"
	parts := strings.SplitN(cred, constants.At, 3)
	if len(parts) != 3 {
		return nil, result.ErrWithCode(result.ErrCodeCredentialsInvalid, "签名格式不正确")
	}
	signatureHex := parts[0]
	timestamp := parts[1]
	bodyHash := parts[2]
	if signatureHex == constants.Empty || timestamp == constants.Empty || bodyHash == constants.Empty {
		return nil, result.ErrWithCode(result.ErrCodeCredentialsInvalid, "签名、时间戳或摘要不能为空")
	}

	principal, secret, err := a.loader.LoadById(appId)
	if err != nil {
		return nil, err
	}
	if principal == nil || secret == constants.Empty {
		return nil, result.ErrWithCode(result.ErrCodeExternalAppNotFound, result.ErrMessageExternalAppNotFound)
	}

	// Recompute signature: hex(HMAC-SHA256(secret, appId + "\n" + timestamp + "\n" + bodyHash))
	var sb strings.Builder
	_, _ = sb.WriteString(appId)
	_ = sb.WriteByte(constants.ByteNewline)
	_, _ = sb.WriteString(timestamp)
	_ = sb.WriteByte(constants.ByteNewline)
	_, _ = sb.WriteString(bodyHash)

	secretBytes, err := hex.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode app secret: %w", err)
	}

	mac := hmac.New(sha256.New, secretBytes)
	_, _ = mac.Write([]byte(sb.String()))
	expectedMac := mac.Sum(nil)
	providedMac, err := hex.DecodeString(signatureHex)
	if err != nil {
		return nil, result.ErrWithCode(result.ErrCodeSignatureInvalid, "签名解码失败")
	}
	if !hmac.Equal(expectedMac, providedMac) {
		return nil, result.ErrSignatureInvalid
	}

	logger.Infof("openapi authentication successful for principal '%s'", principal.Id)
	return principal, nil
}

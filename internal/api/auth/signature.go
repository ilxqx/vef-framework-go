package auth

import (
	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cast"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	isecurity "github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// SignatureStrategy implements api.AuthStrategy for HMAC signature authentication.
// It extracts credentials from HTTP headers and delegates authentication
// to the security.AuthManager, following the Spring Security pattern.
//
// Required headers:
//   - X-App-ID: Application identifier (used as Principal)
//   - X-Timestamp: Unix timestamp in seconds
//   - X-Nonce: Random string for replay attack prevention
//   - X-Signature: HMAC signature in hex encoding
type SignatureStrategy struct {
	authManager security.AuthManager
}

// NewSignature creates a new signature authentication strategy.
// The authManager is used to delegate the actual authentication to SignatureAuthenticator.
func NewSignature(authManager security.AuthManager) api.AuthStrategy {
	return &SignatureStrategy{
		authManager: authManager,
	}
}

// Name returns the strategy name.
func (*SignatureStrategy) Name() string {
	return api.AuthStrategySignature
}

// Authenticate extracts credentials from request headers and delegates
// authentication to the AuthManager.
// Headers are extracted and formatted as: Principal=AppID, Credentials="timestamp:nonce:signature".
func (s *SignatureStrategy) Authenticate(ctx fiber.Ctx, _ map[string]any) (*security.Principal, error) {
	appID := ctx.Get(constants.HeaderXAppID)
	timestampStr := ctx.Get(constants.HeaderXTimestamp)
	nonce := ctx.Get(constants.HeaderXNonce)
	signature := ctx.Get(constants.HeaderXSignature)

	if appID == constants.Empty {
		return nil, result.ErrAppIDRequired
	}

	if timestampStr == constants.Empty {
		return nil, result.ErrTimestampRequired
	}

	if nonce == constants.Empty {
		return nil, result.ErrNonceRequired
	}

	if signature == constants.Empty {
		return nil, result.ErrSignatureRequired
	}

	timestamp, err := cast.ToInt64E(timestampStr)
	if err != nil {
		return nil, result.ErrTimestampInvalid
	}

	credentials := &security.SignatureCredentials{
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: signature,
	}

	authentication := security.Authentication{
		Kind:        isecurity.AuthKindSignature,
		Principal:   appID,
		Credentials: credentials,
	}

	return s.authManager.Authenticate(ctx.Context(), authentication)
}

package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/security"
)

// ExternalApp represents an external application for signature authentication.
type ExternalApp struct {
	ID        string
	Secret    string
	Principal *security.Principal
}

// ExternalAppLoader loads external app by ID.
type ExternalAppLoader interface {
	Load(ctx context.Context, appID string) (*ExternalApp, error)
}

// SignatureStrategy implements api.AuthStrategy for HMAC signature authentication.
type SignatureStrategy struct {
	appLoader ExternalAppLoader
	tolerance time.Duration

	// Header names
	appIDHeader     string
	timestampHeader string
	signatureHeader string
}

// SignatureOption configures SignatureStrategy.
type SignatureOption func(*SignatureStrategy)

// WithTimestampTolerance sets the timestamp tolerance.
func WithTimestampTolerance(d time.Duration) SignatureOption {
	return func(s *SignatureStrategy) {
		s.tolerance = d
	}
}

// WithHeaders sets custom header names.
func WithHeaders(appID, timestamp, signature string) SignatureOption {
	return func(s *SignatureStrategy) {
		s.appIDHeader = appID
		s.timestampHeader = timestamp
		s.signatureHeader = signature
	}
}

// NewSignature creates a new signature authentication strategy.
func NewSignature(loader ExternalAppLoader, opts ...SignatureOption) api.AuthStrategy {
	s := &SignatureStrategy{
		appLoader:       loader,
		tolerance:       5 * time.Minute,
		appIDHeader:     "X-App-ID",
		timestampHeader: "X-Timestamp",
		signatureHeader: "X-Signature",
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Name returns the strategy name.
func (s *SignatureStrategy) Name() string {
	return api.AuthStrategySignature
}

// Authenticate validates the signature and returns the principal.
func (s *SignatureStrategy) Authenticate(ctx fiber.Ctx, _ map[string]any) (*security.Principal, error) {
	appID := ctx.Get(s.appIDHeader)
	timestamp := ctx.Get(s.timestampHeader)
	signature := ctx.Get(s.signatureHeader)

	if appID == constants.Empty || timestamp == constants.Empty || signature == constants.Empty {
		return nil, ErrMissingAuthHeaders
	}

	// Validate timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, ErrInvalidTimestamp
	}

	if time.Since(time.Unix(ts, 0)) > s.tolerance {
		return nil, ErrRequestExpired
	}

	// Load app
	app, err := s.appLoader.Load(ctx.Context(), appID)
	if err != nil {
		return nil, err
	}

	// Verify signature
	if !s.verifySignature(appID, timestamp, ctx.Body(), app.Secret, signature) {
		return nil, ErrInvalidSignature
	}

	return app.Principal, nil
}

// verifySignature verifies the HMAC signature.
func (s *SignatureStrategy) verifySignature(appID, timestamp string, body []byte, secret, signature string) bool {
	message := appID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

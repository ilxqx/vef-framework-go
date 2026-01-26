package auth

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/api/common"
	"github.com/ilxqx/vef-framework-go/security"
)

var defaultTokenExtractor = extractors.Chain(
	extractors.FromAuthHeader(constants.AuthSchemeBearer),
	extractors.FromQuery(constants.QueryKeyAccessToken),
)

// TokenAuthenticator validates a token and returns the principal.
type TokenAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*security.Principal, error)
}

// BearerStrategy implements api.AuthStrategy for Bearer token authentication.
type BearerStrategy struct {
	extractor      extractors.Extractor
	authenticators []TokenAuthenticator
}

// bearerOption configures BearerStrategy.
type bearerOption func(*BearerStrategy)

// WithTokenExtractor sets a custom token extractor.
func WithTokenExtractor(e extractors.Extractor) bearerOption {
	return func(s *BearerStrategy) {
		s.extractor = e
	}
}

// NewBearer creates a new Bearer token authentication strategy.
func NewBearer(authenticators []TokenAuthenticator, opts ...bearerOption) api.AuthStrategy {
	s := &BearerStrategy{
		authenticators: authenticators,
		extractor:      defaultTokenExtractor,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Name returns the strategy name.
func (s *BearerStrategy) Name() string {
	return api.AuthStrategyBearer
}

// Authenticate validates the bearer token and returns the principal.
func (s *BearerStrategy) Authenticate(ctx fiber.Ctx, _ map[string]any) (*security.Principal, error) {
	token, err := s.extractor.Extract(ctx)
	if err != nil {
		extractErr := err
		if errors.Is(err, extractors.ErrNotFound) {
			extractErr = fiber.ErrUnauthorized
		}

		if op := common.Operation(ctx); op != nil {
			return nil, &common.BaseError{
				Identifier: &op.Identifier,
				Err:        extractErr,
			}
		}

		return nil, extractErr
	}

	for _, auth := range s.authenticators {
		principal, err := auth.Authenticate(ctx.Context(), token)
		if err != nil {
			return nil, err
		}

		if principal != nil {
			return principal, nil
		}
	}

	return nil, ErrInvalidToken
}

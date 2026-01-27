package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/api/shared"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

const (
	defaultRateLimitMax        = 100
	defaultRateLimitExpiration = 5 * time.Minute
)

// RateLimit handles rate limiting based on operation config.
// It uses a shared in-memory store to maintain state across requests.
type RateLimit struct {
	h fiber.Handler
}

// NewRateLimit creates a new rate limit middleware with shared state.
func NewRateLimit() api.Middleware {
	return &RateLimit{
		h: limiter.New(limiter.Config{
			LimiterMiddleware: limiter.FixedWindow{},
			MaxFunc: func(ctx fiber.Ctx) int {
				if op := shared.Operation(ctx); op != nil {
					return op.RateLimit.Max
				}

				return defaultRateLimitMax
			},
			Expiration:             defaultRateLimitExpiration,
			SkipFailedRequests:     false,
			SkipSuccessfulRequests: false,
			KeyGenerator: func(ctx fiber.Ctx) string {
				var sb strings.Builder
				if req := shared.Request(ctx); req != nil {
					sb.WriteString(req.Resource)
					sb.WriteByte(constants.ByteColon)
					sb.WriteString(req.Version)
					sb.WriteByte(constants.ByteColon)
					sb.WriteString(req.Action)
					sb.WriteByte(constants.ByteColon)
					sb.WriteString(webhelpers.GetIP(ctx))
					sb.WriteByte(constants.ByteColon)
				}

				principal := contextx.Principal(ctx)
				if principal == nil {
					principal = security.PrincipalAnonymous
				}

				sb.WriteString(principal.ID)

				return sb.String()
			},
			LimitReached: func(fiber.Ctx) error {
				return result.ErrTooManyRequests
			},
		}),
	}
}

// Name returns the middleware name.
func (*RateLimit) Name() string {
	return "ratelimit"
}

// Order returns the middleware order.
func (*RateLimit) Order() int {
	return -70
}

// Process handles the rate limiting.
func (m *RateLimit) Process(ctx fiber.Ctx) error {
	return m.h(ctx)
}

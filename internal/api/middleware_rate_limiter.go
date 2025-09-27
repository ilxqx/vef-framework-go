package api

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/webhelpers"
	"github.com/samber/lo"
)

// buildRateLimiterMiddleware creates a rate limiting middleware.
// It uses a sliding window algorithm and generates keys based on resource, version, action, IP, and user ID.
func buildRateLimiterMiddleware(manager api.Manager) fiber.Handler {
	handler := limiter.New(limiter.Config{
		LimiterMiddleware: limiter.SlidingWindow{},
		MaxFunc: func(ctx fiber.Ctx) int {
			request := contextx.APIRequest(ctx)
			definition := manager.Lookup(request.Identifier)
			return lo.Ternary(definition.HasRateLimit(), definition.Limit.Max, 10)
		},
		Expiration: 2 * time.Minute,
		// ExpirationFunc: func(ctx fiber.Ctx) time.Duration {
		// 	request := contextx.APIRequest(ctx)
		// 	definition := manager.Lookup(request.Identifier)
		// 	return lo.Ternary(definition.HasRateLimit(), definition.RateExpiration, 30*time.Second)
		// },
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		KeyGenerator: func(ctx fiber.Ctx) string {
			request := contextx.APIRequest(ctx)
			var sb strings.Builder
			_, _ = sb.WriteString(request.Resource)
			_ = sb.WriteByte(constants.ByteColon)
			_, _ = sb.WriteString(request.Version)
			_ = sb.WriteByte(constants.ByteColon)
			_, _ = sb.WriteString(request.Action)
			_ = sb.WriteByte(constants.ByteColon)
			_, _ = sb.WriteString(webhelpers.GetIP(ctx))
			_ = sb.WriteByte(constants.ByteColon)

			principal := contextx.Principal(ctx)
			if principal == nil {
				principal = security.PrincipalAnonymous
			}
			_, _ = sb.WriteString(principal.Id)

			return sb.String()
		},
		LimitReached: func(ctx fiber.Ctx) error {
			r := &result.Result{
				Code:    result.ErrCodeTooManyRequests,
				Message: i18n.T(result.ErrMessageTooManyRequests),
			}

			return r.ResponseWithStatus(ctx, fiber.StatusTooManyRequests)
		},
	})

	return func(ctx fiber.Ctx) error {
		return handler(ctx)
	}
}

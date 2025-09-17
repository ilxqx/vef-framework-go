package middleware

import (
	"github.com/gofiber/fiber/v3"
	loggerMiddleware "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/utils"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// newRequestRecordMiddleware returns a middleware that records request metrics.
// It logs IP, latency (ms/μs), and status code and reports structured errors if present.
func newRequestRecordMiddleware() app.Middleware {
	handler := loggerMiddleware.New(loggerMiddleware.Config{
		LoggerFunc: func(ctx fiber.Ctx, data *loggerMiddleware.Data, config loggerMiddleware.Config) error {
			ip := utils.GetIP(ctx)
			latency := data.Stop.Sub(data.Start)

			logger := contextx.Logger(ctx)
			logger.Infof(
				"request completed | ip: %s | latency: %s | status: %d",
				ip,
				lo.TernaryF(latency.Milliseconds() > 0, func() string {
					return cast.ToString(latency.Milliseconds()) + "ms"
				}, func() string {
					return cast.ToString(latency.Microseconds()) + "μs"
				}),
				ctx.Response().StatusCode(),
			)
			if data.ChainErr != nil {
				if err, ok := result.AsErr(data.ChainErr); ok {
					logger.Warnf("request failed with error: %v [%d]", err.Message, err.Code)
				} else {
					logger.Errorf("request failed with error: %v", data.ChainErr)
				}
			}
			return nil
		},
	})

	return &SimpleMiddleware{
		handler: handler,
		name:    "request_record",
		order:   -100,
	}
}

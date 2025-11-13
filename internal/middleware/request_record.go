package middleware

import (
	"path"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cast"

	loggerMiddleware "github.com/gofiber/fiber/v3/middleware/logger"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/middleware"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// simplifyUserAgent reduces verbose UA strings to concise "Client/OS" format for log readability.
func simplifyUserAgent(ua string) string {
	if ua == constants.Empty {
		return "Unknown"
	}

	ua = strings.ToLower(ua)

	var os string
	switch {
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		os = "iOS"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macintosh"):
		os = "Mac"
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	default:
		os = "Unknown"
	}

	var client string
	switch {
	case strings.Contains(ua, "micromessenger"):
		client = "WeChat"
	case strings.Contains(ua, "dingtalk"):
		client = "DingTalk"
	case strings.Contains(ua, "alipay"):
		client = "Alipay"
	case strings.Contains(ua, "edg/") || strings.Contains(ua, "edge/"):
		client = "Edge"
	case strings.Contains(ua, "chrome/") && !strings.Contains(ua, "edg"):
		client = "Chrome"
	case strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome"):
		client = "Safari"
	case strings.Contains(ua, "firefox/"):
		client = "Firefox"
	case strings.Contains(ua, "postman"):
		client = "Postman"
	case strings.Contains(ua, "curl"):
		client = "cURL"
	case strings.Contains(ua, "okhttp"):
		client = "OkHttp"
	default:
		client = constants.Empty
	}

	if client != constants.Empty && os != "Unknown" {
		return client + "/" + os
	} else if client != constants.Empty {
		return client
	}

	return os
}

func isSpaStaticRequest(ctx fiber.Ctx, spaConfigs []*middleware.SpaConfig) bool {
	if ctx.Method() != fiber.MethodGet {
		return false
	}

	reqPath := ctx.Path()
	for _, config := range spaConfigs {
		spaPath := config.Path
		if spaPath == constants.Empty {
			spaPath = constants.Slash
		}

		staticPath := path.Join(spaPath, "static/")
		if reqPath == spaPath || strings.HasPrefix(reqPath, staticPath) {
			return true
		}
	}

	return false
}

// NewRequestRecordMiddleware skips SPA static assets to reduce log noise while capturing API traffic.
func NewRequestRecordMiddleware(spaConfigs []*middleware.SpaConfig) app.Middleware {
	handler := loggerMiddleware.New(loggerMiddleware.Config{
		Next: func(ctx fiber.Ctx) bool {
			return isSpaStaticRequest(ctx, spaConfigs)
		},
		LoggerFunc: func(ctx fiber.Ctx, data *loggerMiddleware.Data, config loggerMiddleware.Config) error {
			method, path := ctx.Method(), ctx.Path()
			ip, latency, status := webhelpers.GetIp(ctx), data.Stop.Sub(data.Start), ctx.Response().StatusCode()
			ua := simplifyUserAgent(ctx.Get(fiber.HeaderUserAgent))

			var latencyStr string
			if ms := latency.Milliseconds(); ms > 0 {
				latencyStr = cast.ToString(ms) + "ms"
			} else {
				latencyStr = cast.ToString(latency.Microseconds()) + "μs"
			}

			logger := contextx.Logger(ctx)
			logger.Infof(
				"Request completed | %s %s | ip: %s | ua: %s | latency: %s | status: %d",
				method,
				path,
				ip,
				ua,
				latencyStr,
				status,
			)

			if data.ChainErr != nil {
				if err, ok := result.AsErr(data.ChainErr); ok {
					logger.Warnf("Request failed with error: %v [%d]", err.Message, err.Code)
				} else {
					logger.Errorf("Request failed with error: %v", data.ChainErr)
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

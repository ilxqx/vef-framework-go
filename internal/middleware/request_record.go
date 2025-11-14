package middleware

import (
	"path"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/muesli/termenv"
	"github.com/spf13/cast"

	loggerMiddleware "github.com/gofiber/fiber/v3/middleware/logger"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/middleware"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

var (
	output              = termenv.DefaultOutput()
	labelValueSeparator = " | "
	ipLabel             = output.String("ip: ").Foreground(termenv.ANSIBrightBlack).String()
	uaLabel             = output.String("ua: ").Foreground(termenv.ANSIBrightBlack).String()
	latencyLabel        = output.String("latency: ").Foreground(termenv.ANSIBrightBlack).String()
	statusLabel         = output.String("status: ").Foreground(termenv.ANSIBrightBlack).String()
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
		return client + constants.Slash + os
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

func formatRequestDetails(ctx fiber.Ctx, data *loggerMiddleware.Data) string {
	method, path := ctx.Method(), ctx.Path()
	ip, latency, status := webhelpers.GetIp(ctx), data.Stop.Sub(data.Start), ctx.Response().StatusCode()
	ua := simplifyUserAgent(ctx.Get(fiber.HeaderUserAgent))

	var latencyStr string
	if ms := latency.Milliseconds(); ms > 0 {
		latencyStr = cast.ToString(ms) + "ms"
	} else {
		latencyStr = cast.ToString(latency.Microseconds()) + "μs"
	}

	var sb strings.Builder

	_, _ = sb.WriteString(labelValueSeparator)
	_, _ = sb.WriteString(
		output.String(method).Foreground(termenv.ANSIBrightCyan).String(),
	)
	_, _ = sb.WriteString(constants.Space)
	_, _ = sb.WriteString(
		output.String(path).Foreground(termenv.ANSIBrightCyan).String(),
	)

	_, _ = sb.WriteString(labelValueSeparator)
	_, _ = sb.WriteString(ipLabel)
	_, _ = sb.WriteString(
		output.String(ip).Foreground(termenv.ANSIBrightCyan).String(),
	)

	_, _ = sb.WriteString(labelValueSeparator)
	_, _ = sb.WriteString(uaLabel)
	_, _ = sb.WriteString(
		output.String(ua).Foreground(termenv.ANSIBrightCyan).String(),
	)

	_, _ = sb.WriteString(labelValueSeparator)
	_, _ = sb.WriteString(latencyLabel)
	ms := latency.Milliseconds()

	var latencyOutput termenv.Style
	switch {
	case ms >= 1000:
		latencyOutput = output.String(latencyStr).Foreground(termenv.ANSIBrightRed).Bold()
	case ms >= 500:
		latencyOutput = output.String(latencyStr).Foreground(termenv.ANSIBrightYellow).Bold()
	case ms >= 200:
		latencyOutput = output.String(latencyStr).Foreground(termenv.ANSIBrightBlue)
	default:
		latencyOutput = output.String(latencyStr).Foreground(termenv.ANSIBrightGreen)
	}

	_, _ = sb.WriteString(latencyOutput.String())

	_, _ = sb.WriteString(labelValueSeparator)
	_, _ = sb.WriteString(statusLabel)

	statusColor := termenv.ANSIBrightRed
	if status >= 200 && status < 300 {
		statusColor = termenv.ANSIBrightGreen
	}

	_, _ = sb.WriteString(
		output.String(cast.ToString(status)).Foreground(statusColor).String(),
	)

	return sb.String()
}

// NewRequestRecordMiddleware skips SPA static assets to reduce log noise while capturing API traffic.
func NewRequestRecordMiddleware(spaConfigs []*middleware.SpaConfig) app.Middleware {
	handler := loggerMiddleware.New(loggerMiddleware.Config{
		Next: func(ctx fiber.Ctx) bool {
			return isSpaStaticRequest(ctx, spaConfigs)
		},
		LoggerFunc: func(ctx fiber.Ctx, data *loggerMiddleware.Data, _ loggerMiddleware.Config) error {
			details := formatRequestDetails(ctx, data)
			logger := contextx.Logger(ctx)

			if data.ChainErr != nil {
				if err, ok := result.AsErr(data.ChainErr); ok {
					var sb strings.Builder

					_, _ = sb.WriteString("Request completed with error: ")
					_, _ = sb.WriteString(data.ChainErr.Error())
					_ = sb.WriteByte(constants.ByteLeftParenthesis)
					_, _ = sb.WriteString(cast.ToString(err.Code))
					_ = sb.WriteByte(constants.ByteRightParenthesis)

					logger.Warnf(
						"%s%s",
						output.String(sb.String()).Foreground(termenv.ANSIBrightYellow).String(),
						details,
					)
				} else {
					var sb strings.Builder

					_, _ = sb.WriteString("Request failed with error: ")
					_, _ = sb.WriteString(data.ChainErr.Error())

					logger.Errorf(
						"%s%s",
						output.String(sb.String()).Foreground(termenv.ANSIBrightRed).String(),
						details,
					)
				}
			} else {
				logger.Infof(
					"%s%s",
					output.String("Request completed").Foreground(termenv.ANSIBrightGreen).String(),
					details,
				)
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

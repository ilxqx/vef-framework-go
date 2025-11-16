package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/middleware"
)

type spaMiddleware struct {
	configs []*middleware.SpaConfig
}

func (*spaMiddleware) Name() string {
	return "spa"
}

func (*spaMiddleware) Order() int {
	return 1000
}

func (s *spaMiddleware) Apply(router fiber.Router) {
	for _, config := range s.configs {
		applySpa(router, config)
	}

	router.Use(func(ctx fiber.Ctx) error {
		if ctx.Method() == fiber.MethodGet {
			path := ctx.Path()
			for _, config := range s.configs {
				if strings.HasPrefix(path, config.Path) {
					ctx.Path(config.Path)

					return ctx.RestartRouting()
				}
			}
		}

		return ctx.Next()
	})
}

func applySpa(router fiber.Router, config *middleware.SpaConfig) {
	group := router.Group(
		config.Path,
		etag.New(etag.Config{
			Weak: true,
		}),
		helmet.New(helmet.Config{
			XFrameOptions:             "sameorigin",
			ReferrerPolicy:            "no-referrer",
			XSSProtection:             "1; mode=block",
			CrossOriginEmbedderPolicy: "require-corp",
			CrossOriginOpenerPolicy:   "same-origin-allow-popups",
			CrossOriginResourcePolicy: "same-origin",
			ContentSecurityPolicy:     "default-src 'self'; img-src * data:; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		}),
	)

	group.Get("/", static.New("index.html", static.Config{
		FS:            config.FS,
		Browse:        false,
		Download:      false,
		CacheDuration: 30 * time.Second,
		MaxAge:        0,
		Compress:      true,
	}))

	group.Get("/static/*", static.New(constants.Empty, static.Config{
		FS:            config.FS,
		Browse:        false,
		Download:      false,
		CacheDuration: 10 * time.Minute,
		MaxAge:        int((8 * time.Hour).Seconds()),
		Compress:      true,
		NotFoundHandler: func(ctx fiber.Ctx) error {
			ctx.Path(lo.Ternary(config.Path == constants.Empty, constants.Slash, config.Path))

			return ctx.RestartRouting()
		},
	}))
}

func NewSpaMiddleware(configs []*middleware.SpaConfig) app.Middleware {
	if len(configs) == 0 {
		return nil
	}

	for _, config := range configs {
		if config.Path == constants.Empty {
			config.Path = constants.Slash
		}
	}

	return &spaMiddleware{
		configs: configs,
	}
}

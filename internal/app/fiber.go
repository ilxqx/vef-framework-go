package app

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/api"
)

// createFiberApp creates a new Fiber application with the given configuration.
// It parses the body limit, sets up custom context, and configures various Fiber settings
// including timeouts, encoders, validators, and error handlers.
func createFiberApp(cfg *config.AppConfig) (*fiber.App, error) {
	specifiedLimit := strings.TrimSpace(cfg.BodyLimit)

	bodyLimit, err := humanize.ParseBytes(
		lo.Ternary(
			specifiedLimit != constants.Empty,
			specifiedLimit,
			"10mib",
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body limit: %w", err)
	}

	return fiber.NewWithCustomCtx(
		func(app *fiber.App) fiber.CustomCtx {
			return &CustomCtx{
				DefaultCtx: *fiber.NewDefaultCtx(app),
			}
		},
		fiber.Config{
			AppName:         lo.CoalesceOrEmpty(cfg.Name, constants.VEFName+"-app"),
			BodyLimit:       int(bodyLimit),
			CaseSensitive:   true,
			IdleTimeout:     30 * time.Second,
			ErrorHandler:    handleError,
			JSONEncoder:     json.Marshal,
			JSONDecoder:     json.Unmarshal,
			StrictRouting:   false,
			StructValidator: newStructValidator(),
			ServerHeader:    constants.VEFName,
			Concurrency:     1024 * 1024,
			ReadBufferSize:  8192,
			WriteBufferSize: 8192,
			Immutable:       false,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    120 * time.Second,
		},
	), nil
}

// configureFiberApp configures the Fiber application with middlewares and routes.
// Middlewares are separated into before (order < 0) and after (order > 0) groups,
// sorted by order, and applied around the API engine registration.
// This ensures proper middleware execution order relative to route handlers.
func configureFiberApp(
	app *fiber.App,
	middlewares []Middleware,
	apiEngine api.Engine,
	openApiEngine api.Engine,
) {
	// Separate middlewares into before and after groups based on order
	beforeMiddlewares := lo.Filter(middlewares, func(mid Middleware, _ int) bool {
		return mid != nil && mid.Order() < 0
	})
	afterMiddlewares := lo.Filter(middlewares, func(mid Middleware, _ int) bool {
		return mid != nil && mid.Order() > 0
	})

	// Sort middlewares by order ascending
	slices.SortFunc(beforeMiddlewares, func(a, b Middleware) int {
		return a.Order() - b.Order()
	})
	slices.SortFunc(afterMiddlewares, func(a, b Middleware) int {
		return a.Order() - b.Order()
	})

	// Apply before middlewares
	for _, mid := range beforeMiddlewares {
		logger.Infof("Applying before middleware '%s'", mid.Name())
		mid.Apply(app)
	}

	// Connect API engines
	apiEngine.Connect(app)
	openApiEngine.Connect(app)

	// Apply after middlewares
	for _, mid := range afterMiddlewares {
		logger.Infof("Applying after middleware '%s'", mid.Name())
		mid.Apply(app)
	}
}

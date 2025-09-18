package app

import (
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/samber/lo"
)

func createFiberApp(config *config.AppConfig) *fiber.App {
	specifiedLimit := strings.TrimSpace(config.BodyLimit)
	bodyLimit, err := humanize.ParseBytes(
		lo.Ternary(
			specifiedLimit != constants.Empty,
			specifiedLimit,
			"10mib",
		),
	)
	if err != nil {
		logger.Errorf("Failed to parse body limit: %v", err)
	}

	return fiber.NewWithCustomCtx(
		func(app *fiber.App) fiber.CustomCtx {
			return &CustomCtx{
				DefaultCtx: *fiber.NewDefaultCtx(app),
			}
		},
		fiber.Config{
			AppName:         config.Name,
			BodyLimit:       int(bodyLimit),
			CaseSensitive:   true,
			IdleTimeout:     30 * time.Second,
			ErrorHandler:    handleError,
			JSONEncoder:     json.Marshal,
			JSONDecoder:     json.Unmarshal,
			StrictRouting:   false,
			StructValidator: newStructValidator(),
			ServerHeader:    "vef",
			Concurrency:     1024 * 1024,
			ReadBufferSize:  8192,
			WriteBufferSize: 8192,
			Immutable:       false,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    120 * time.Second,
		},
	)
}

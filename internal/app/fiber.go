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
	specifiedLimit := strings.TrimSpace(config.BodyLimit) // specifiedLimit gets the configured body limit
	bodyLimit, err := humanize.ParseBytes(                // bodyLimit parses the body limit string to bytes
		lo.Ternary(
			specifiedLimit != constants.Empty,
			specifiedLimit,
			"10mib", // Default body limit is 10 MiB
		),
	)
	if err != nil {
		logger.Errorf("Failed to parse body limit: %v", err)
	}

	return fiber.NewWithCustomCtx( // NewWithCustomCtx creates Fiber app with custom context
		func(app *fiber.App) fiber.CustomCtx { // Custom context factory function
			return &CustomCtx{
				DefaultCtx: *fiber.NewDefaultCtx(app), // DefaultCtx wraps the default Fiber context
			}
		},
		fiber.Config{ // Config sets up Fiber application configuration
			AppName:         config.Name,          // AppName sets the application name
			BodyLimit:       int(bodyLimit),       // BodyLimit sets the maximum request body size
			CaseSensitive:   true,                 // CaseSensitive enables case-sensitive routing
			IdleTimeout:     30 * time.Second,     // IdleTimeout sets connection idle timeout
			ErrorHandler:    handleError,          // ErrorHandler sets the global error handler
			JSONEncoder:     json.Marshal,         // JSONEncoder sets the JSON encoder
			JSONDecoder:     json.Unmarshal,       // JSONDecoder sets the JSON decoder
			StrictRouting:   false,                // StrictRouting disables strict routing
			StructValidator: newStructValidator(), // StructValidator sets the struct validator
			ServerHeader:    "vef",                // ServerHeader sets the server header
			Concurrency:     1024 * 1024,          // Concurrency sets the maximum number of concurrent connections
			ReadBufferSize:  8192,                 // ReadBufferSize sets the read buffer size
			WriteBufferSize: 8192,                 // WriteBufferSize sets the write buffer size
			Immutable:       false,                // Immutable disables immutable mode
			ReadTimeout:     30 * time.Second,     // ReadTimeout sets the read timeout
			WriteTimeout:    120 * time.Second,    // WriteTimeout sets the write timeout
		},
	)
}

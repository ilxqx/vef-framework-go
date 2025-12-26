package vef

import (
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/ilxqx/vef-framework-go/internal/api"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/config"
	"github.com/ilxqx/vef-framework-go/internal/cron"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/event"
	ilog "github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/internal/mcp"
	"github.com/ilxqx/vef-framework-go/internal/middleware"
	"github.com/ilxqx/vef-framework-go/internal/mold"
	"github.com/ilxqx/vef-framework-go/internal/monitor"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/internal/redis"
	"github.com/ilxqx/vef-framework-go/internal/schema"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/internal/storage"
	"github.com/ilxqx/vef-framework-go/log"
)

// Default timeout for framework startup and shutdown.
const defaultTimeout = 30 * time.Second

func newFxLogger() fxevent.Logger {
	return &fxevent.SlogLogger{
		Logger: ilog.NewSLogger("vef", 5, log.LevelWarn),
	}
}

// Run starts the VEF framework with the provided options.
// It initializes all core modules and runs the application.
func Run(options ...fx.Option) {
	// Core framework modules
	opts := []fx.Option{
		fx.WithLogger(newFxLogger),
		config.Module,
		database.Module,
		orm.Module,
		middleware.Module,
		api.Module,
		security.Module,
		event.Module,
		cron.Module,
		redis.Module,
		mold.Module,
		storage.Module,
		schema.Module,
		monitor.Module,
		mcp.Module,
		app.Module,
	}

	opts = append(opts, options...)
	opts = append(
		opts,
		fx.Invoke(startApp),
		fx.StartTimeout(defaultTimeout),
		fx.StopTimeout(defaultTimeout*2),
	)

	app := fx.New(opts...)
	app.Run()
}

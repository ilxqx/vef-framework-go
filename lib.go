package vef

import (
	"time"

	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/internal/api"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/cache"
	"github.com/ilxqx/vef-framework-go/internal/config"
	"github.com/ilxqx/vef-framework-go/internal/cron"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/event"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/internal/middleware"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/internal/redis"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/internal/trans"
	logPkg "github.com/ilxqx/vef-framework-go/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Default timeout for framework startup and shutdown
const defaultTimeout = 30 * time.Second

func newFxLogger() fxevent.Logger {
	return &fxevent.SlogLogger{
		Logger: log.NewSLogger("vef", 5, logPkg.LevelWarn),
	}
}

// Run starts the VEF framework with the provided options.
// It initializes all core modules and runs the application.
func Run(options ...fx.Option) {
	// Core framework modules in dependency order
	frameworkOptions := []fx.Option{
		fx.WithLogger(newFxLogger),
		config.Module,
		database.Module,
		orm.Module,
		middleware.Module,
		api.Module,
		security.Module,
		cache.Module,
		event.Module,
		cron.Module,
		redis.Module,
		trans.Module,
		apis.Module,
		app.Module,
	}

	frameworkOptions = append(frameworkOptions, options...)
	frameworkOptions = append(
		frameworkOptions,
		fx.Invoke(start),
		fx.StartTimeout(defaultTimeout),
		fx.StopTimeout(defaultTimeout*2),
	)

	app := fx.New(frameworkOptions...)
	app.Run()
}

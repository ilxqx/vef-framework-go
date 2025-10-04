package vef

import (
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/api"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/cache"
	"github.com/ilxqx/vef-framework-go/internal/config"
	"github.com/ilxqx/vef-framework-go/internal/cron"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/event"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/internal/middleware"
	"github.com/ilxqx/vef-framework-go/internal/mold"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/internal/redis"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/internal/storage"
	logPkg "github.com/ilxqx/vef-framework-go/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Default timeout for framework startup and shutdown
const defaultTimeout = 30 * time.Second

func newFxLogger() fxevent.Logger {
	return &fxevent.SlogLogger{
		Logger: log.NewSLogger(constants.VEFName, 5, logPkg.LevelWarn),
	}
}

// Run starts the VEF framework with the provided options.
// It initializes all core modules and runs the application.
func Run(options ...fx.Option) {
	// Core framework modules in dependency order
	opts := []fx.Option{
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
		mold.Module,
		storage.Module,
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

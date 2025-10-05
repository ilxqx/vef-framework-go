package test

import (
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	configPkg "github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/api"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/cache"
	"github.com/ilxqx/vef-framework-go/internal/config"
	"github.com/ilxqx/vef-framework-go/internal/cron"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/event"
	"github.com/ilxqx/vef-framework-go/internal/middleware"
	"github.com/ilxqx/vef-framework-go/internal/mold"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/internal/redis"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/internal/storage"
)

// MockConfig implements config.Config for testing without file dependencies.
type MockConfig struct{}

func (m *MockConfig) Unmarshal(key string, target any) error {
	// No-op for testing - we'll use fx.Replace for specific configs
	return nil
}

// NewTestApp creates a new test application with Fx dependency injection.
// It returns the Fx app instance and the VEF application.
func NewTestApp(t testing.TB, options ...fx.Option) (*app.App, func()) {
	var testApp *app.App

	// Build fx options
	opts := []fx.Option{
		fx.NopLogger,
		// Replace configs - must replace config.Config to avoid file reading
		fx.Replace(
			fx.Annotate(
				&MockConfig{},
				fx.As(new(configPkg.Config)),
			),
			&configPkg.AppConfig{
				Name:      "test-app",
				Port:      0, // Random port
				BodyLimit: "100mib",
			},
		),
		// Core framework modules
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

	// Add additional modules
	if len(options) > 0 {
		opts = append(opts, options...)
	}

	// Populate app
	opts = append(opts, fx.Populate(&testApp))

	fxApp := fxtest.New(t, opts...)
	fxApp.RequireStart()

	return testApp, func() {
		fxApp.RequireStop()
	}
}

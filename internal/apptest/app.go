package apptest

import (
	"context"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/api"
	"github.com/ilxqx/vef-framework-go/internal/app"
	iconfig "github.com/ilxqx/vef-framework-go/internal/config"
	"github.com/ilxqx/vef-framework-go/internal/cron"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/event"
	"github.com/ilxqx/vef-framework-go/internal/mcp"
	"github.com/ilxqx/vef-framework-go/internal/middleware"
	"github.com/ilxqx/vef-framework-go/internal/mold"
	"github.com/ilxqx/vef-framework-go/internal/monitor"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/internal/redis"
	"github.com/ilxqx/vef-framework-go/internal/schema"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/internal/storage"
)

// MockConfig implements config.Config for testing without file dependencies.
type MockConfig struct{}

func (*MockConfig) Unmarshal(_ string, _ any) error {
	return nil
}

// NewTestApp creates a test application with Fx dependency injection.
// Returns the app instance and a cleanup function.
func NewTestApp(t testing.TB, options ...fx.Option) (*app.App, func()) {
	var testApp *app.App

	opts := append(buildOptions(options...), fx.Populate(&testApp))
	fxApp := fxtest.New(t, opts...)
	fxApp.RequireStart()

	return testApp, fxApp.RequireStop
}

// NewTestAppWithErr creates a test application and returns any startup errors.
// Useful for testing error conditions during app initialization.
func NewTestAppWithErr(t testing.TB, options ...fx.Option) (*app.App, func(), error) {
	var testApp *app.App

	opts := append(buildOptions(options...), fx.Populate(&testApp))
	fxApp := fx.New(opts...)

	startCtx, cancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
	defer cancel()

	err := fxApp.Start(startCtx)
	cleanup := createCleanupFunc(t, fxApp)

	return testApp, cleanup, err
}

func createCleanupFunc(t testing.TB, fxApp *fx.App) func() {
	return func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
		defer cancel()

		if err := fxApp.Stop(stopCtx); err != nil {
			t.Logf("Failed to stop app: %v", err)
		}
	}
}

func buildOptions(options ...fx.Option) []fx.Option {
	opts := []fx.Option{
		fx.NopLogger,
		fx.Replace(
			fx.Annotate(&MockConfig{}, fx.As(new(config.Config))),
			&config.AppConfig{
				Name:      "test-app",
				Port:      0,
				BodyLimit: "100mib",
			},
		),
		iconfig.Module,
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
		monitor.Module,
		schema.Module,
		mcp.Module,
		app.Module,
	}

	return append(opts, options...)
}

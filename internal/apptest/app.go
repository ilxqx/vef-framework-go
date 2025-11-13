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
	"github.com/ilxqx/vef-framework-go/internal/middleware"
	"github.com/ilxqx/vef-framework-go/internal/mold"
	"github.com/ilxqx/vef-framework-go/internal/monitor"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/internal/redis"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/internal/storage"
)

const testTimeout = fx.DefaultTimeout

// MockConfig implements config.Config for testing without file dependencies.
type MockConfig struct{}

func (m *MockConfig) Unmarshal(key string, target any) error {
	return nil
}

// NewTestApp creates a new test application with Fx dependency injection.
// It returns the Fx app instance and the VEF application.
func NewTestApp(t testing.TB, options ...fx.Option) (*app.App, func()) {
	var testApp *app.App

	opts := buildOptions(options...)
	// Populate app
	opts = append(opts, fx.Populate(&testApp))

	fxApp := fxtest.New(t, opts...)
	fxApp.RequireStart()

	return testApp, func() {
		fxApp.RequireStop()
	}
}

// NewTestAppWithErr creates a new test application and returns any startup errors.
// Unlike NewTestApp, this function does not require the app to start successfully.
// It's useful for testing error conditions during app initialization.
func NewTestAppWithErr(t testing.TB, options ...fx.Option) (*app.App, func(), error) {
	var testApp *app.App

	opts := buildOptions(options...)
	// Populate app
	opts = append(opts, fx.Populate(&testApp))

	// Use fx.New instead of fxtest.New to allow capturing errors
	fxApp := fx.New(opts...)

	// Try to start and capture any error
	startCtx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	if err := fxApp.Start(startCtx); err != nil {
		return testApp, func() {
			stopCtx, stopCancel := context.WithTimeout(context.Background(), testTimeout)
			defer stopCancel()

			if err := fxApp.Stop(stopCtx); err != nil {
				t.Logf("Failed to stop app: %v", err)
			}
		}, err
	}

	return testApp, func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), testTimeout)
		defer stopCancel()

		if err := fxApp.Stop(stopCtx); err != nil {
			t.Logf("Failed to stop app: %v", err)
		}
	}, nil
}

func buildOptions(options ...fx.Option) []fx.Option {
	// Build fx options
	opts := []fx.Option{
		fx.NopLogger,
		// Replace configs - must replace config.Config to avoid file reading
		fx.Replace(
			fx.Annotate(
				&MockConfig{},
				fx.As(new(config.Config)),
			),
			&config.AppConfig{
				Name:      "test-app",
				Port:      0, // Random port
				BodyLimit: "100mib",
			},
		),
		// Core framework modules
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
		app.Module,
	}

	// Add additional modules
	if len(options) > 0 {
		opts = append(opts, options...)
	}

	return opts
}

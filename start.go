package vef

import (
	"github.com/ilxqx/vef-framework-go/internal/app"
	"go.uber.org/fx"
)

// startParams contains the dependencies required for application startup.
type startParams struct {
	fx.In
	App         *app.App
	Middlewares []app.Middleware `group:"vef:app:middlewares"`
}

// start initializes and starts the application with the provided middleware chain.
// It registers the application stop hook with the fx lifecycle manager.
func start(lc fx.Lifecycle, params startParams) error {
	params.App.Use(params.Middlewares...)

	if err := <-params.App.Start(); err != nil {
		return err
	}

	lc.Append(fx.StopHook(params.App.Stop))
	return nil
}

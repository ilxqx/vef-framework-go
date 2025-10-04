package vef

import (
	"github.com/ilxqx/vef-framework-go/internal/app"
	"go.uber.org/fx"
)

// startApp starts the application.
// It registers the application stop hook with the fx lifecycle manager.
func startApp(lc fx.Lifecycle, app *app.App) error {
	if err := <-app.Start(); err != nil {
		return err
	}

	lc.Append(fx.StopHook(app.Stop))
	return nil
}

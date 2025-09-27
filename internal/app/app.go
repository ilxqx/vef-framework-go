package app

import (
	"fmt"
	"slices"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/api"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/muesli/termenv"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

var logger = log.Named("app")

type App struct {
	app           *fiber.App
	port          uint16
	middlewares   []Middleware
	apiEngine     api.Engine
	openApiEngine api.Engine
}

func (a *App) Unwrap() *fiber.App {
	return a.app
}

func (a *App) Use(middlewares ...Middleware) {
	a.middlewares = append(a.middlewares, middlewares...)
}

func (a *App) Start() <-chan error {
	logger.Info("Starting VEF application...")

	// errChan is a buffered channel for error communication
	errChan := make(chan error, 1)
	go func() {
		if err := a.configure(); err != nil {
			errChan <- err
			return
		}

		if err := a.app.Listen(
			fmt.Sprintf(":%d", a.port),
			fiber.ListenConfig{
				EnablePrintRoutes: false,
				ShutdownTimeout:   30 * time.Second,
				BeforeServeFunc: func(*fiber.App) error {
					errChan <- nil

					output := termenv.DefaultOutput()
					fmt.Printf(` _    ______________
| |  / / ____/ ____/
| | / / __/ / /_    
| |/ / /___/ __/    
|___/_____/_/                   %s
--------------------------------------------------
`, output.String(constants.VEFVersion).Foreground(termenv.ANSIBrightGreen).String())

					logger.Infof("VEF application started successfully on port %d", a.port)
					return nil
				},
			},
		); err != nil {
			logger.Errorf("Failed to start VEF application: %v", err)
			errChan <- err
		}
	}()

	return errChan
}

func (a *App) configure() error {
	beforeMiddlewares := lo.Filter(a.middlewares, func(mid Middleware, _ int) bool { // beforeMiddlewares filters middlewares with negative order
		return mid != nil && mid.Order() < 0
	})
	afterMiddlewares := lo.Filter(a.middlewares, func(mid Middleware, _ int) bool { // afterMiddlewares filters middlewares with positive order
		return mid != nil && mid.Order() > 0
	})
	// SortFunc sorts before middlewares by order
	slices.SortFunc(
		beforeMiddlewares,
		func(a, b Middleware) int {
			// Sort by order ascending
			return a.Order() - b.Order()
		},
	)
	// SortFunc sorts after middlewares by order
	slices.SortFunc(
		afterMiddlewares,
		func(a, b Middleware) int {
			// Sort by order ascending
			return a.Order() - b.Order()
		},
	)

	for _, mid := range beforeMiddlewares {
		logger.Infof("Applying before middleware '%s'", mid.Name())
		mid.Apply(a.app)
	}

	a.apiEngine.Connect(a.app)
	a.openApiEngine.Connect(a.app)

	for _, mid := range afterMiddlewares {
		logger.Infof("Applying after middleware '%s'", mid.Name())
		mid.Apply(a.app)
	}

	return nil
}

func (a *App) Stop() error {
	logger.Info("Stopping VEF application...")
	return a.app.ShutdownWithTimeout(time.Second * 30)
}

type AppParams struct {
	fx.In
	Config        *config.AppConfig
	ApiEngine     api.Engine `name:"vef:api:engine"`
	OpenApiEngine api.Engine `name:"vef:openapi:engine"`
}

func New(params AppParams) *App {
	logger.Info("Initializing VEF application...")
	return &App{
		app:           createFiberApp(params.Config),
		port:          params.Config.Port,
		apiEngine:     params.ApiEngine,
		openApiEngine: params.OpenApiEngine,
	}
}

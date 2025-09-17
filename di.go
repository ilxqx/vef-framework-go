package vef

import (
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/middleware"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

var Provide = fx.Provide
var Supply = fx.Supply
var Annotate = fx.Annotate
var As = fx.As
var ParamTags = fx.ParamTags
var ResultTags = fx.ResultTags
var Self = fx.Self
var Invoke = fx.Invoke
var Decorate = fx.Decorate
var Module = fx.Module
var Private = fx.Private
var OnStart = fx.OnStart
var OnStop = fx.OnStop

type Hook = fx.Hook
type HookFunc = fx.HookFunc

var From = fx.From
var Replace = fx.Replace
var Populate = fx.Populate

type In = fx.In
type Out = fx.Out
type Lifecycle = fx.Lifecycle

func StartHook[T HookFunc](start T) Hook {
	return fx.StartHook(start)
}

func StopHook[T HookFunc](stop T) Hook {
	return fx.StopHook(stop)
}

func StartStopHook[T1 HookFunc, T2 HookFunc](start T1, stop T2) Hook {
	return fx.StartStopHook(start, stop)
}

// ProvideAPIResource provides an API resource to the dependency injection container.
// The resource will be registered in the "vef:api:resources" group.
func ProvideAPIResource(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.As(new(api.Resource)),
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	)
}

// ProvideOpenAPIResource provides an OpenAPI resource to the dependency injection container.
// The resource will be registered in the "vef:openapi:resources" group.
func ProvideOpenAPIResource(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.As(new(api.Resource)),
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:openapi:resources"`),
		),
	)
}

// ProvideMiddleware provides a middleware to the dependency injection container.
// The middleware will be registered in the "vef:app:middlewares" group.
func ProvideMiddleware(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.ParamTags(paramTags...),
			fx.As(new(app.Middleware)),
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
	)
}

// ProvideSPAConfig provides a Single Page Application configuration to the dependency injection container.
// The config will be registered in the "vef:spa" group.
func ProvideSPAConfig(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:spa"`),
		),
	)
}

// SupplySPAConfigs supplies multiple Single Page Application configurations to the dependency injection container.
// All configs will be registered in the "vef:spa" group.
func SupplySPAConfigs(config *middleware.SPAConfig, configs ...*middleware.SPAConfig) fx.Option {
	spaConfigs := make([]any, 0, len(configs)+1)
	spaConfigs = append(
		spaConfigs,
		fx.Annotate(
			config,
			fx.ResultTags(`group:"vef:spa"`),
		),
	)
	if len(configs) > 0 {
		spaConfigs = append(
			spaConfigs,
			lo.Map(configs, func(item *middleware.SPAConfig, _ int) any {
				return fx.Annotate(
					item,
					fx.ResultTags(`group:"vef:spa"`),
				)
			}),
		)
	}

	return fx.Supply(spaConfigs...)
}

package vef

import (
	"github.com/samber/lo"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/mcp"
	"github.com/ilxqx/vef-framework-go/middleware"
)

var (
	Provide    = fx.Provide
	Supply     = fx.Supply
	Annotate   = fx.Annotate
	As         = fx.As
	ParamTags  = fx.ParamTags
	ResultTags = fx.ResultTags
	Self       = fx.Self
	Invoke     = fx.Invoke
	Decorate   = fx.Decorate
	Module     = fx.Module
	Private    = fx.Private
	OnStart    = fx.OnStart
	OnStop     = fx.OnStop
)

type (
	Hook     = fx.Hook
	HookFunc = fx.HookFunc
)

var (
	From     = fx.From
	Replace  = fx.Replace
	Populate = fx.Populate
)

type (
	In        = fx.In
	Out       = fx.Out
	Lifecycle = fx.Lifecycle
)

func StartHook[T HookFunc](start T) Hook {
	return fx.StartHook(start)
}

func StopHook[T HookFunc](stop T) Hook {
	return fx.StopHook(stop)
}

func StartStopHook[T1, T2 HookFunc](start T1, stop T2) Hook {
	return fx.StartStopHook(start, stop)
}

// ProvideApiResource provides an Api resource to the dependency injection container.
// The resource will be registered in the "vef:api:resources" group.
func ProvideApiResource(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.As(new(api.Resource)),
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	)
}

// ProvideOpenApiResource provides an OpenApi resource to the dependency injection container.
// The resource will be registered in the "vef:openapi:resources" group.
func ProvideOpenApiResource(constructor any, paramTags ...string) fx.Option {
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

// ProvideSpaConfig provides a Single Page Application configuration to the dependency injection container.
// The config will be registered in the "vef:spa" group.
func ProvideSpaConfig(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:spa"`),
		),
	)
}

// SupplySpaConfigs supplies multiple Single Page Application configurations to the dependency injection container.
// All configs will be registered in the "vef:spa" group.
func SupplySpaConfigs(config *middleware.SpaConfig, configs ...*middleware.SpaConfig) fx.Option {
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
			lo.Map(configs, func(item *middleware.SpaConfig, _ int) any {
				return fx.Annotate(
					item,
					fx.ResultTags(`group:"vef:spa"`),
				)
			}),
		)
	}

	return fx.Supply(spaConfigs...)
}

// ProvideMcpTools provides an MCP tool provider.
func ProvideMcpTools(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.As(new(mcp.ToolProvider)),
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:mcp:tools"`),
		),
	)
}

// ProvideMcpResources provides an MCP resource provider.
func ProvideMcpResources(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.As(new(mcp.ResourceProvider)),
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:mcp:resources"`),
		),
	)
}

// ProvideMcpResourceTemplates provides an MCP resource template provider.
func ProvideMcpResourceTemplates(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.As(new(mcp.ResourceTemplateProvider)),
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:mcp:templates"`),
		),
	)
}

// ProvideMcpPrompts provides an MCP prompt provider.
func ProvideMcpPrompts(constructor any, paramTags ...string) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.As(new(mcp.PromptProvider)),
			fx.ParamTags(paramTags...),
			fx.ResultTags(`group:"vef:mcp:prompts"`),
		),
	)
}

// SupplyMcpServerInfo supplies MCP server info.
func SupplyMcpServerInfo(info *mcp.ServerInfo) fx.Option {
	return fx.Supply(info)
}

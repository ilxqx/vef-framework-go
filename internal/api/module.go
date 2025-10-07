package api

import "go.uber.org/fx"

var Module = fx.Module(
	"vef:api",
	fx.Provide(
		fx.Annotate(
			NewHandlerParamResolverManager,
			fx.ParamTags(`group:"vef:api:param_resolvers"`),
		),
		fx.Private,
	),
	fx.Provide(
		// provide managers
		fx.Annotate(
			NewManager,
			fx.ParamTags(`group:"vef:api:resources"`, `optional:"true"`),
			fx.ResultTags(`name:"vef:api:manager"`),
		),
		fx.Annotate(
			NewManager,
			fx.ParamTags(`group:"vef:openapi:resources"`, `optional:"true"`),
			fx.ResultTags(`name:"vef:openapi:manager"`),
		),

		// provide policies
		fx.Annotate(
			NewDefaultApiPolicy,
			fx.ResultTags(`name:"vef:api:policy"`),
		),
		fx.Annotate(
			NewOpenApiPolicy,
			fx.ResultTags(`name:"vef:openapi:policy"`),
		),

		// provide engines
		fx.Annotate(
			NewEngine,
			fx.ParamTags(`name:"vef:api:manager"`, `name:"vef:api:policy"`, `optional:"true"`),
			fx.ResultTags(`name:"vef:api:engine"`),
		),
		fx.Annotate(
			NewEngine,
			fx.ParamTags(`name:"vef:openapi:manager"`, `name:"vef:openapi:policy"`, `optional:"true"`),
			fx.ResultTags(`name:"vef:openapi:engine"`),
		),
	),
)

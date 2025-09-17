package api

import "go.uber.org/fx"

var Module = fx.Module(
	"vef:api",
	fx.Provide(
		fx.Annotate(
			newParamResolver,
			fx.ParamTags(`group:"vef:api:param_resolvers"`),
		),
		fx.Private,
	),
	fx.Provide(
		fx.Annotate(
			newManager,
			fx.ParamTags(`group:"vef:api:resources"`),
			fx.ResultTags(`name:"vef:api:manager"`),
		),
		fx.Annotate(
			newManager,
			fx.ParamTags(`group:"vef:openapi:resources"`),
			fx.ResultTags(`name:"vef:openapi:manager"`),
		),
		// provide policies
		fx.Annotate(
			newDefaultApiPolicy,
			fx.ResultTags(`name:"vef:api:policy"`),
		),
		fx.Annotate(
			newOpenApiPolicy,
			fx.ResultTags(`name:"vef:openapi:policy"`),
		),
		// provide engines using the same constructor with different named params
		fx.Annotate(
			newEngine,
			fx.ParamTags(`name:"vef:api:manager"`, `name:"vef:api:policy"`),
			fx.ResultTags(`name:"vef:api:engine"`),
		),
		fx.Annotate(
			newEngine,
			fx.ParamTags(`name:"vef:openapi:manager"`, `name:"vef:openapi:policy"`),
			fx.ResultTags(`name:"vef:openapi:engine"`),
		),
	),
)

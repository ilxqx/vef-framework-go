package api

import "go.uber.org/fx"

var Module = fx.Module(
	"vef:api",
	fx.Provide(
		fx.Annotate(
			NewHandlerParamResolverManager,
			fx.ParamTags(`group:"vef:api:handler_param_resolvers"`),
		),
		fx.Annotate(
			NewFactoryParamResolverManager,
			fx.ParamTags(`group:"vef:api:factory_param_resolvers"`),
		),
		fx.Private,
	),
	fx.Provide(
		fx.Annotate(
			NewManager,
			fx.ParamTags(`group:"vef:api:resources"`),
			fx.ResultTags(`name:"vef:api:manager"`),
		),
		fx.Annotate(
			NewManager,
			fx.ParamTags(`group:"vef:openapi:resources"`),
			fx.ResultTags(`name:"vef:openapi:manager"`),
		),
		fx.Annotate(
			NewDefaultApiPolicy,
			fx.ResultTags(`name:"vef:api:policy"`),
		),
		fx.Annotate(
			NewOpenApiPolicy,
			fx.ResultTags(`name:"vef:openapi:policy"`),
		),
		fx.Annotate(
			NewEngine,
			fx.ParamTags(
				`name:"vef:api:manager"`,
				`name:"vef:api:policy"`,
				`optional:"true"`, // PermissionChecker
				`optional:"true"`, // DataPermissionResolver
				``,                // orm.Db
				``,                // event.Publisher
			),
			fx.ResultTags(`name:"vef:api:engine"`),
		),
		fx.Annotate(
			NewEngine,
			fx.ParamTags(
				`name:"vef:openapi:manager"`,
				`name:"vef:openapi:policy"`,
				`optional:"true"`, // PermissionChecker
				`optional:"true"`, // DataPermissionResolver
				``,                // orm.Db
				``,                // event.Publisher
			),
			fx.ResultTags(`name:"vef:openapi:engine"`),
		),
	),
)

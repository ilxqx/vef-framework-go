package orm

import "go.uber.org/fx"

// Module provides the Orm functionality for the VEF framework.
// It registers the database provider and logs initialization status.
var Module = fx.Module(
	"vef:orm",
	fx.Provide(
		New,
		fx.Annotate(
			NewDbFactoryParamResolver,
			fx.ResultTags(`group:"vef:api:factory_param_resolvers"`),
		),
	),
)

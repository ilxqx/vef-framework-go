package middleware

import "go.uber.org/fx"

var Module = fx.Module(
	"vef:middleware",
	fx.Provide(
		fx.Annotate(
			newRequestIdMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
		fx.Annotate(
			newLoggerMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
		fx.Annotate(
			newRecoveryMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
		fx.Annotate(
			newRequestRecordMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
		fx.Annotate(
			newCorsMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
		fx.Annotate(
			newContentTypeMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
		fx.Annotate(
			newCompressionMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
		fx.Annotate(
			newHeadersMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
		fx.Annotate(
			newSPAMiddleware,
			fx.ParamTags(`group:"vef:spa"`),
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
	),
)

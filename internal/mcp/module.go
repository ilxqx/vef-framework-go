package mcp

import (
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/internal/mcp/prompts"
	"github.com/ilxqx/vef-framework-go/internal/mcp/tools"
)

var Module = fx.Module(
	"vef:mcp",
	fx.Provide(
		NewServer,
		NewHandler,
		fx.Annotate(
			NewMiddleware,
			fx.ResultTags(`group:"vef:app:middlewares"`),
		),
	),
	tools.Module,
	prompts.Module,
)

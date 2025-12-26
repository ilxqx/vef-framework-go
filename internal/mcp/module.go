package mcp

import (
	"github.com/ilxqx/vef-framework-go/internal/mcp/prompts"
	"github.com/ilxqx/vef-framework-go/internal/mcp/tools"
	"go.uber.org/fx"
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

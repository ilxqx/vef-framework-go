package result

import "go.uber.org/fx"

var Module = fx.Module(
	"vef:result",
	fx.Invoke(initErrors),
)

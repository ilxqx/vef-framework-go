package security

import (
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

var logger = log.Named("security")

// Module provides the security-related dependencies for the application.
var Module = fx.Module(
	"vef:security",
	fx.Provide(
		// Provide JWT instance
		fx.Annotate(
			func(config *config.AppConfig) (*security.JWT, error) {
				return security.NewJWT(&security.JWTConfig{
					Audience: lo.SnakeCase(config.Name),
				})
			},
		),
		// Provide JWT authenticator
		fx.Annotate(
			NewJWTAuthenticator,
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide JWT refresh authenticator
		fx.Annotate(
			NewJWTRefreshAuthenticator,
			fx.ParamTags(``, `optional:"true"`),
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide JWT token generator
		NewJWTTokenGenerator,
		// Provide OpenAPI authenticator (requires ExternalAppLoader implementation from user)
		fx.Annotate(
			NewOpenAPIAuthenticator,
			fx.ParamTags(`optional:"true"`),
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide Password authenticator (requires UserLoader implementation from user)
		fx.Annotate(
			NewPasswordAuthenticator,
			fx.ParamTags(`optional:"true"`, `optional:"true"`),
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide authentication manager
		fx.Annotate(
			NewAuthManager,
			fx.ParamTags(`group:"vef:security:authenticators"`),
		),
		// Provide auth resource
		fx.Annotate(
			NewAuthResource,
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	),
	fx.Invoke(func() {
		logger.Info("Security module initialized")
	}),
)

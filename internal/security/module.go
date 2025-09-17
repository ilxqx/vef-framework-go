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
			func(config *config.AppConfig) *security.JWT {
				return security.NewJWT(&security.JWTConfig{
					Audience: lo.SnakeCase(config.Name),
				})
			},
		),
		// Provide JWT authenticator
		fx.Annotate(
			newJWTAuthenticator,
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide JWT refresh authenticator
		fx.Annotate(
			newJWTRefreshAuthenticator,
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide JWT token generator
		newJWTTokenGenerator,
		// Provide OpenAPI authenticator (requires ExternalAppLoader implementation from user)
		fx.Annotate(
			newOpenAPIAuthenticator,
			fx.ParamTags(`optional:"true"`),
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide Password authenticator (requires UserLoader implementation from user)
		fx.Annotate(
			newPasswordAuthenticator,
			fx.ParamTags(`optional:"true"`),
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide authentication manager
		fx.Annotate(
			newAuthManager,
			fx.ParamTags(`group:"vef:security:authenticators"`),
		),
		// Provide auth resource
		fx.Annotate(
			newAuthResource,
			fx.ParamTags(`optional:"true"`),
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	),
	fx.Invoke(func() {
		logger.Info("Security module initialized")
	}),
)

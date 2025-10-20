package security

import (
	"github.com/samber/lo"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/security"
)

var logger = log.Named("security")

// Module provides the security-related dependencies for the application.
var Module = fx.Module(
	"vef:security",
	fx.Provide(
		// Provide Jwt instance
		fx.Annotate(
			func(config *config.AppConfig) (*security.Jwt, error) {
				return security.NewJwt(&security.JwtConfig{
					Audience: lo.SnakeCase(config.Name),
				})
			},
		),
		// Provide Jwt authenticator
		fx.Annotate(
			NewJwtAuthenticator,
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide Jwt refresh authenticator
		fx.Annotate(
			NewJwtRefreshAuthenticator,
			fx.ParamTags(``, `optional:"true"`),
			fx.ResultTags(`group:"vef:security:authenticators"`),
		),
		// Provide Jwt token generator
		NewJwtTokenGenerator,
		// Provide OpenApi authenticator (requires ExternalAppLoader implementation from user)
		fx.Annotate(
			NewOpenApiAuthenticator,
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
		// Provide RBAC permission checker (requires RolePermissionsLoader implementation from user)
		fx.Annotate(
			NewRbacPermissionChecker,
			fx.ParamTags(`optional:"true"`),
		),
		// Provide RBAC data permission resolver (requires RolePermissionsLoader implementation from user)
		fx.Annotate(
			NewRbacDataPermResolver,
			fx.ParamTags(`optional:"true"`),
		),
		// Provide auth resource
		fx.Annotate(
			NewAuthResource,
			fx.ParamTags(``, ``, `optional:"true"`),
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	),
)

package security

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// NewAuthResource creates a new authentication resource with the provided auth manager and token generator.
func NewAuthResource(authManager security.AuthManager, tokenGenerator security.TokenGenerator, userInfoLoader security.UserInfoLoader, publisher event.Publisher) api.Resource {
	return &AuthResource{
		authManager:    authManager,
		tokenGenerator: tokenGenerator,
		userInfoLoader: userInfoLoader,
		publisher:      publisher,
		Resource: api.NewRPCResource(
			"security/auth",
			api.WithOperations(
				api.OperationSpec{
					Action:    "login",
					Public:    true,
					RateLimit: loginRateLimit,
				},
				api.OperationSpec{
					Action:    "refresh",
					Public:    true,
					RateLimit: refreshRateLimit,
				},
				api.OperationSpec{
					Action: "logout",
				},
				api.OperationSpec{
					Action: "get_user_info",
				},
			),
		),
	}
}

// AuthResource handles authentication-related Api endpoints.
type AuthResource struct {
	api.Resource

	authManager    security.AuthManager
	tokenGenerator security.TokenGenerator
	userInfoLoader security.UserInfoLoader
	publisher      event.Publisher
}

// LoginParams represents the request parameters for user login.
type LoginParams struct {
	api.P

	// Authentication contains user credentials
	security.Authentication
}

// Login authenticates a user and returns token credentials.
func (a *AuthResource) Login(ctx fiber.Ctx, params LoginParams) error {
	loginIp := webhelpers.GetIP(ctx)
	userAgent := ctx.Get(fiber.HeaderUserAgent)
	traceID := contextx.RequestID(ctx)
	username := params.Principal

	principal, err := a.authManager.Authenticate(ctx.Context(), params.Authentication)
	if err != nil {
		var (
			failReason string
			errorCode  int
		)

		if resErr, ok := result.AsErr(err); ok {
			failReason = resErr.Message
			errorCode = resErr.Code
		} else {
			failReason = err.Error()
			errorCode = result.ErrCodeUnknown
		}

		loginEvent := security.NewLoginEvent(
			params.Kind,
			constants.Empty,
			username,
			loginIp,
			userAgent,
			traceID,
			false,
			failReason,
			errorCode,
		)
		a.publisher.Publish(loginEvent)

		return err
	}

	credentials, err := a.tokenGenerator.Generate(principal)
	if err != nil {
		return err
	}

	loginEvent := security.NewLoginEvent(
		params.Kind,
		principal.ID,
		username,
		loginIp,
		userAgent,
		traceID,
		true,
		constants.Empty,
		0,
	)
	a.publisher.Publish(loginEvent)

	return result.Ok(credentials).Response(ctx)
}

// RefreshParams represents the request parameters for token refresh operation.
type RefreshParams struct {
	api.P

	RefreshToken string `json:"refreshToken"`
}

// Refresh refreshes the access token using a valid refresh token.
// User data reload logic is handled by JwtRefreshAuthenticator.
func (a *AuthResource) Refresh(ctx fiber.Ctx, params RefreshParams) error {
	principal, err := a.authManager.Authenticate(ctx.Context(), security.Authentication{
		Kind:      AuthKindRefresh,
		Principal: params.RefreshToken,
	})
	if err != nil {
		return err
	}

	credentials, err := a.tokenGenerator.Generate(principal)
	if err != nil {
		return err
	}

	return result.Ok(credentials).Response(ctx)
}

// Logout returns success immediately.
// Token invalidation should be handled on the client side by removing stored tokens.
func (*AuthResource) Logout(ctx fiber.Ctx) error {
	return result.Ok().Response(ctx)
}

// GetUserInfo retrieves user information via UserInfoLoader.
// Requires a UserInfoLoader implementation to be provided.
func (a *AuthResource) GetUserInfo(ctx fiber.Ctx, principal *security.Principal, params api.Params) error {
	if a.userInfoLoader == nil {
		return result.ErrNotImplemented(i18n.T(result.ErrMessageUserInfoLoaderNotImplemented))
	}

	userInfo, err := a.userInfoLoader.LoadUserInfo(ctx.Context(), principal, params)
	if err != nil {
		return err
	}

	return result.Ok(userInfo).Response(ctx)
}

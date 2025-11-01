package security

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// NewAuthResource creates a new authentication resource with the provided auth manager and token generator.
func NewAuthResource(authManager security.AuthManager, tokenGenerator security.TokenGenerator, userInfoLoader security.UserInfoLoader) api.Resource {
	return &AuthResource{
		authManager:    authManager,
		tokenGenerator: tokenGenerator,
		userInfoLoader: userInfoLoader,
		Resource: api.NewResource(
			"security/auth",
			api.WithApis(
				api.Spec{
					Action: "login",
					Public: true,
				},
				api.Spec{
					Action: "refresh",
					Public: true,
					Limit: api.RateLimit{
						Max: 1,
					},
				},
				api.Spec{
					Action: "logout",
				},
				api.Spec{
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
}

// LoginParams represents the request parameters for user login.
type LoginParams struct {
	api.P

	// Authentication contains user credentials
	security.Authentication
}

// Login authenticates a user and returns token credentials.
// It validates the provided credentials and generates access tokens upon successful authentication.
func (a *AuthResource) Login(ctx fiber.Ctx, params LoginParams) error {
	// Authenticate user credentials using the auth manager
	principal, err := a.authManager.Authenticate(ctx.Context(), params.Authentication)
	if err != nil {
		return err
	}

	// Generate tokens for the authenticated user
	credentials, err := a.tokenGenerator.Generate(principal)
	if err != nil {
		return err
	}

	// Return the generated credentials as a successful response
	return result.Ok(credentials).Response(ctx)
}

// RefreshParams represents the request parameters for token refresh operation.
type RefreshParams struct {
	api.P

	// RefreshToken is the Jwt refresh token used to generate new access tokens
	RefreshToken string `json:"refreshToken"`
}

// Refresh refreshes the access token using a valid refresh token.
// It validates the refresh token and generates new access tokens.
// Note: The user data reload logic is now handled by JwtRefreshAuthenticator.
func (a *AuthResource) Refresh(ctx fiber.Ctx, params RefreshParams) error {
	// Validate and extract/reload user information from the refresh token
	principal, err := a.authManager.Authenticate(ctx.Context(), security.Authentication{
		Type:      AuthTypeRefresh,
		Principal: params.RefreshToken,
	})
	if err != nil {
		return err
	}

	// Generate new access and refresh tokens for the authenticated user
	credentials, err := a.tokenGenerator.Generate(principal)
	if err != nil {
		return err
	}

	// Return the generated credentials as a successful response
	return result.Ok(credentials).Response(ctx)
}

// Logout logs out the authenticated user and invalidates their session.
// This is a client-side logout implementation that returns success immediately.
// Token invalidation should be handled on the client side by removing stored tokens.
func (a *AuthResource) Logout(ctx fiber.Ctx) error {
	return result.Ok().Response(ctx)
}

// GetUserInfo retrieves detailed information about the currently authenticated user.
// It requires a UserInfoLoader implementation to be provided.
func (a *AuthResource) GetUserInfo(ctx fiber.Ctx, principal *security.Principal) error {
	if a.userInfoLoader == nil {
		return result.Err(i18n.T("user_info_loader_not_implemented"), result.WithCode(result.ErrCodeNotImplemented))
	}

	// Get API request from context to extract params
	var params map[string]any
	if req := contextx.ApiRequest(ctx); req != nil {
		params = req.Params
	}

	if params == nil {
		params = make(map[string]any)
	}

	// Load user information using the UserInfoLoader
	userInfo, err := a.userInfoLoader.LoadUserInfo(ctx.Context(), principal, params)
	if err != nil {
		return err
	}

	return result.Ok(userInfo).Response(ctx)
}

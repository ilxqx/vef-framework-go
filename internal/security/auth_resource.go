package security

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// NewAuthResource creates a new authentication resource with the provided auth manager and token generator.
func NewAuthResource(authManager security.AuthManager, tokenGenerator security.TokenGenerator) api.Resource {
	return &AuthResource{
		authManager:    authManager,
		tokenGenerator: tokenGenerator,
		Resource: api.NewResource(
			"security/auth",
			api.WithAPIs(
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
			),
		),
	}
}

// AuthResource handles authentication-related API endpoints.
type AuthResource struct {
	api.Resource

	authManager    security.AuthManager
	tokenGenerator security.TokenGenerator
}

// LoginParams represents the request parameters for user login.
type LoginParams struct {
	api.In

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
	api.In

	// RefreshToken is the JWT refresh token used to generate new access tokens
	RefreshToken string `json:"refreshToken"`
}

// Refresh refreshes the access token using a valid refresh token.
// It validates the refresh token and generates new access tokens.
// Note: The user data reload logic is now handled by JWTRefreshAuthenticator.
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

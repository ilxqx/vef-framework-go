package security

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// NewAuthResource creates a new authentication resource with the provided auth manager, token generator, and user loader.
func NewAuthResource(userLoader security.UserLoader, authManager security.AuthManager, tokenGenerator security.TokenGenerator) api.Resource {
	return &authResource{
		authManager:    authManager,
		tokenGenerator: tokenGenerator,
		userLoader:     userLoader,
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

// authResource handles authentication-related API endpoints.
type authResource struct {
	api.Resource
	authManager    security.AuthManager    // authManager handles user authentication operations
	tokenGenerator security.TokenGenerator // tokenGenerator creates JWT tokens for authenticated users
	userLoader     security.UserLoader     // userLoader loads user principal by id
}

// loginParams represents the request parameters for user login.
type loginParams struct {
	api.In
	security.Authentication // Authentication contains user credentials
}

// Login authenticates a user and returns token credentials.
// It validates the provided credentials and generates access tokens upon successful authentication.
func (a *authResource) Login(ctx fiber.Ctx, params loginParams) error {
	// Authenticate user credentials using the auth manager
	principal, err := a.authManager.Authenticate(params.Authentication)
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

// refreshParams represents the request parameters for token refresh operation.
type refreshParams struct {
	api.In
	RefreshToken string `json:"refreshToken"` // RefreshToken is the JWT refresh token used to generate new access tokens
}

// Refresh refreshes the access token using a valid refresh token.
// It validates the refresh token, reloads the user principal, and generates new access tokens.
func (a *authResource) Refresh(ctx fiber.Ctx, params refreshParams) error {
	if a.userLoader == nil {
		return result.ErrWithCode(result.ErrCodeNotImplemented, "user loader implementation is required for token refresh")
	}

	// Validate and extract user information from the refresh token
	principal, err := a.authManager.Authenticate(security.Authentication{
		Type:      AuthTypeJWTRefresh,
		Principal: params.RefreshToken,
	})
	if err != nil {
		return err
	}

	// Reload the latest user principal data by ID to ensure current user state
	principal, err = a.userLoader.LoadById(principal.Id)
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
func (a *authResource) Logout(ctx fiber.Ctx) error {
	return result.Ok().Response(ctx)
}

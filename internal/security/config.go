//go:build !test

package security

import "github.com/ilxqx/vef-framework-go/api"

// Production configuration for security module.

// refreshTokenNotBefore controls the not-before duration for refresh tokens in production.
// Using accessTokenExpires/2 prevents immediate refresh token reuse after access token issue.
const refreshTokenNotBefore = accessTokenExpires / 2

// Rate limits for authentication endpoints in production environment.
// Strict limits protect against brute-force attacks and token abuse.
var (
	loginRateLimit = api.RateLimit{
		Max: 5,
	}
	refreshRateLimit = api.RateLimit{
		Max: 1,
	}
)

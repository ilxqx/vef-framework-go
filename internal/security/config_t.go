//go:build test

package security

import (
	"time"

	"github.com/ilxqx/vef-framework-go/api"
)

// Test configuration for security module.

// refreshTokenNotBefore controls the not-before duration for refresh tokens in test environment.
// Using 0 allows immediate token usage in tests.
const refreshTokenNotBefore = time.Duration(0)

// Rate limits for authentication endpoints in test environment.
// High limits prevent test failures from rate limiting.
var (
	loginRateLimit = api.RateLimit{
		Max: 100,
	}
	refreshRateLimit = api.RateLimit{
		Max: 100,
	}
)

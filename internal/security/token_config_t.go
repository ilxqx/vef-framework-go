//go:build test

package security

import "time"

// refreshTokenNotBefore controls the not-before duration for refresh tokens in test environment.
// Using 0 allows immediate token usage in tests.
const refreshTokenNotBefore = time.Duration(0)

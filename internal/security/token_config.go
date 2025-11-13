//go:build !test

package security

// refreshTokenNotBefore controls the not-before duration for refresh tokens in production.
// Using accessTokenExpires/2 prevents immediate refresh token reuse after access token issue.
const refreshTokenNotBefore = accessTokenExpires / 2

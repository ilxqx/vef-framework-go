package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/result"
)

func TestNewJWT(t *testing.T) {
	t.Run("valid hex secret", func(t *testing.T) {
		config := &JWTConfig{
			Secret:   "af6675678bd81ad7c93c4a51d122ef61e9750fe5d42ceac1c33b293f36bc14c2",
			Audience: "test_app",
		}
		jwt, err := NewJWT(config)
		require.NoError(t, err)
		assert.NotNil(t, jwt)
		assert.Equal(t, "test_app", jwt.config.Audience)
	})

	t.Run("invalid hex secret", func(t *testing.T) {
		config := &JWTConfig{
			Secret: "invalid-hex",
		}
		jwt, err := NewJWT(config)
		assert.Error(t, err)
		assert.Nil(t, jwt)
		assert.Contains(t, err.Error(), "failed to decode jwt secret")
	})

	t.Run("empty secret uses default", func(t *testing.T) {
		config := &JWTConfig{
			Secret: "",
		}
		jwt, err := NewJWT(config)
		require.NoError(t, err)
		assert.NotNil(t, jwt)
		assert.Equal(t, 32, len(jwt.secret)) // Default secret is 64 hex chars = 32 bytes
	})

	t.Run("empty audience uses default", func(t *testing.T) {
		config := &JWTConfig{
			Secret:   "af6675678bd81ad7c93c4a51d122ef61e9750fe5d42ceac1c33b293f36bc14c2",
			Audience: "",
		}
		jwt, err := NewJWT(config)
		require.NoError(t, err)
		assert.Equal(t, defaultJwtAudience, jwt.config.Audience)
	})
}

func TestJWTGenerate(t *testing.T) {
	config := &JWTConfig{
		Secret:   "af6675678bd81ad7c93c4a51d122ef61e9750fe5d42ceac1c33b293f36bc14c2",
		Audience: "test_app",
	}
	jwt, err := NewJWT(config)
	require.NoError(t, err)

	t.Run("generate valid token", func(t *testing.T) {
		builder := NewJWTClaimsBuilder().
			WithClaim("user_id", "123").
			WithClaim("username", "testuser")

		token, err := jwt.Generate(builder, 1*time.Hour, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token can be parsed
		claims, err := jwt.Parse(token)
		require.NoError(t, err)
		assert.Equal(t, "123", claims.Claim("user_id"))
		assert.Equal(t, "testuser", claims.Claim("username"))
	})

	t.Run("generate token with not before", func(t *testing.T) {
		builder := NewJWTClaimsBuilder().WithClaim("test", "value")

		// Set nbf to 2 minutes in future (beyond the 1 minute leeway)
		token, err := jwt.Generate(builder, 1*time.Hour, 2*time.Minute)
		require.NoError(t, err)

		// Token should not be valid yet due to nbf
		_, err = jwt.Parse(token)
		assert.ErrorIs(t, err, result.ErrTokenNotValidYet)
	})

	t.Run("standard claims are set correctly", func(t *testing.T) {
		builder := NewJWTClaimsBuilder()
		token, err := jwt.Generate(builder, 1*time.Hour, 0)
		require.NoError(t, err)

		claims, err := jwt.Parse(token)
		require.NoError(t, err)

		assert.Equal(t, jwtIssuer, claims.Claim(claimIssuer))
		assert.Equal(t, "test_app", claims.Claim(claimAudience))
		iat, ok := claims.Claim(claimIssuedAt).(float64)
		require.True(t, ok)
		exp, ok := claims.Claim(claimExpiresAt).(float64)
		require.True(t, ok)
		assert.Greater(t, int64(iat), int64(0))
		assert.Greater(t, int64(exp), int64(iat))
	})
}

func TestJWTParse(t *testing.T) {
	config := &JWTConfig{
		Secret:   "af6675678bd81ad7c93c4a51d122ef61e9750fe5d42ceac1c33b293f36bc14c2",
		Audience: "test_app",
	}
	jwt, err := NewJWT(config)
	require.NoError(t, err)

	t.Run("parse valid token", func(t *testing.T) {
		builder := NewJWTClaimsBuilder().
			WithClaim("user_id", "456").
			WithClaim("role", "admin")

		token, err := jwt.Generate(builder, 1*time.Hour, 0)
		require.NoError(t, err)

		claims, err := jwt.Parse(token)
		require.NoError(t, err)
		assert.Equal(t, "456", claims.Claim("user_id"))
		assert.Equal(t, "admin", claims.Claim("role"))
	})

	t.Run("parse expired token", func(t *testing.T) {
		builder := NewJWTClaimsBuilder().WithClaim("test", "value")
		token, err := jwt.Generate(builder, -1*time.Hour, 0) // Already expired
		require.NoError(t, err)

		_, err = jwt.Parse(token)
		assert.ErrorIs(t, err, result.ErrTokenExpired)
	})

	t.Run("parse token with wrong audience", func(t *testing.T) {
		wrongConfig := &JWTConfig{
			Secret:   "af6675678bd81ad7c93c4a51d122ef61e9750fe5d42ceac1c33b293f36bc14c2",
			Audience: "wrong_app",
		}
		wrongJWT, err := NewJWT(wrongConfig)
		require.NoError(t, err)

		builder := NewJWTClaimsBuilder().WithClaim("test", "value")
		token, err := wrongJWT.Generate(builder, 1*time.Hour, 0)
		require.NoError(t, err)

		// Try to parse with original JWT (different audience)
		_, err = jwt.Parse(token)
		assert.ErrorIs(t, err, result.ErrTokenInvalidAudience)
	})

	t.Run("parse token with wrong secret", func(t *testing.T) {
		wrongConfig := &JWTConfig{
			Secret:   "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Audience: "test_app",
		}
		wrongJWT, err := NewJWT(wrongConfig)
		require.NoError(t, err)

		builder := NewJWTClaimsBuilder().WithClaim("test", "value")
		token, err := wrongJWT.Generate(builder, 1*time.Hour, 0)
		require.NoError(t, err)

		// Try to parse with original JWT (different secret)
		_, err = jwt.Parse(token)
		assert.ErrorIs(t, err, result.ErrTokenInvalid)
	})

	t.Run("parse malformed token", func(t *testing.T) {
		_, err := jwt.Parse("malformed.token.string")
		assert.ErrorIs(t, err, result.ErrTokenInvalid)
	})

	t.Run("parse empty token", func(t *testing.T) {
		_, err := jwt.Parse("")
		assert.ErrorIs(t, err, result.ErrTokenInvalid)
	})
}

func TestJWTErrorMapping(t *testing.T) {
	config := &JWTConfig{
		Secret:   "af6675678bd81ad7c93c4a51d122ef61e9750fe5d42ceac1c33b293f36bc14c2",
		Audience: "test_app",
	}
	jwt, err := NewJWT(config)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		tokenGen      func() string
		expectedError error
	}{
		{
			name: "expired token",
			tokenGen: func() string {
				builder := NewJWTClaimsBuilder()
				token, _ := jwt.Generate(builder, -1*time.Hour, 0)

				return token
			},
			expectedError: result.ErrTokenExpired,
		},
		{
			name: "not yet valid token",
			tokenGen: func() string {
				builder := NewJWTClaimsBuilder()
				token, _ := jwt.Generate(builder, 1*time.Hour, 2*time.Minute)

				return token
			},
			expectedError: result.ErrTokenNotValidYet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.tokenGen()
			_, err := jwt.Parse(token)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestJWTClaimsBuilder(t *testing.T) {
	t.Run("build claims with various types", func(t *testing.T) {
		builder := NewJWTClaimsBuilder().
			WithClaim("string_val", "test").
			WithClaim("int_val", 123).
			WithClaim("bool_val", true).
			WithClaim("float_val", 3.14).
			WithClaim("map_val", map[string]any{"key": "value"})

		claims := builder.build()
		assert.Equal(t, "test", claims["string_val"])
		assert.Equal(t, 123, claims["int_val"])
		assert.Equal(t, true, claims["bool_val"])
		assert.Equal(t, 3.14, claims["float_val"])
		assert.Equal(t, map[string]any{"key": "value"}, claims["map_val"])
	})

	t.Run("overwrite existing claim", func(t *testing.T) {
		builder := NewJWTClaimsBuilder().
			WithClaim("key", "value1").
			WithClaim("key", "value2")

		claims := builder.build()
		assert.Equal(t, "value2", claims["key"])
	})

	t.Run("use specialized claim methods", func(t *testing.T) {
		builder := NewJWTClaimsBuilder().
			WithId("jwt123").
			WithSubject("user456").
			WithType("access").
			WithRoles([]string{"admin", "user"}).
			WithDetails(map[string]any{"email": "test@example.com"})

		id, ok := builder.Id()
		assert.True(t, ok)
		assert.Equal(t, "jwt123", id)

		subject, ok := builder.Subject()
		assert.True(t, ok)
		assert.Equal(t, "user456", subject)

		typ, ok := builder.Type()
		assert.True(t, ok)
		assert.Equal(t, "access", typ)

		roles, ok := builder.Roles()
		assert.True(t, ok)
		assert.Equal(t, []string{"admin", "user"}, roles)

		details, ok := builder.Details()
		assert.True(t, ok)
		assert.Equal(t, map[string]any{"email": "test@example.com"}, details)
	})
}

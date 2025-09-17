package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Run("hashes password successfully", func(t *testing.T) {
		password := "testpassword123"

		hashedPassword, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)
	})

	t.Run("different passwords produce different hashes", func(t *testing.T) {
		password1 := "password123"
		password2 := "password456"

		hash1, err1 := HashPassword(password1)
		hash2, err2 := HashPassword(password2)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("same password produces different hashes due to salt", func(t *testing.T) {
		password := "samepassword"

		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("handles empty password", func(t *testing.T) {
		password := ""

		hashedPassword, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
	})

	t.Run("handles long password", func(t *testing.T) {
		password := strings.Repeat("a", 1000)

		hashedPassword, err := HashPassword(password)

		// bcrypt has a 72-byte limit, so this should error
		assert.Error(t, err)
		assert.Empty(t, hashedPassword)
	})

	t.Run("handles special characters", func(t *testing.T) {
		password := "!@#$%^&*()_+-=[]{}|;:,.<>?"

		hashedPassword, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)
	})

	t.Run("handles unicode characters", func(t *testing.T) {
		password := "密码测试123🔒"

		hashedPassword, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)
	})

	t.Run("hash format is bcrypt", func(t *testing.T) {
		password := "testpassword"

		hashedPassword, err := HashPassword(password)

		require.NoError(t, err)
		// bcrypt hashes start with $2a$, $2b$, $2x$, or $2y$
		assert.True(t, strings.HasPrefix(hashedPassword, "$2"))
		// bcrypt hashes are typically 60 characters long
		assert.Len(t, hashedPassword, 60)
	})
}

func TestVerifyPassword(t *testing.T) {
	t.Run("verifies correct password", func(t *testing.T) {
		password := "testpassword123"
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)

		result := VerifyPassword(password, hashedPassword)

		assert.True(t, result)
	})

	t.Run("rejects incorrect password", func(t *testing.T) {
		correctPassword := "correctpassword"
		incorrectPassword := "incorrectpassword"
		hashedPassword, err := HashPassword(correctPassword)
		require.NoError(t, err)

		result := VerifyPassword(incorrectPassword, hashedPassword)

		assert.False(t, result)
	})

	t.Run("handles empty password", func(t *testing.T) {
		password := ""
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)

		result := VerifyPassword(password, hashedPassword)

		assert.True(t, result)
	})

	t.Run("rejects empty password against non-empty hash", func(t *testing.T) {
		originalPassword := "nonemptypassword"
		hashedPassword, err := HashPassword(originalPassword)
		require.NoError(t, err)

		result := VerifyPassword("", hashedPassword)

		assert.False(t, result)
	})

	t.Run("handles long password verification", func(t *testing.T) {
		password := strings.Repeat("a", 1000)
		hashedPassword, err := HashPassword(password)

		// Since HashPassword should error for long passwords, we expect an error
		assert.Error(t, err)
		assert.Empty(t, hashedPassword)
	})

	t.Run("handles special characters verification", func(t *testing.T) {
		password := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)

		result := VerifyPassword(password, hashedPassword)

		assert.True(t, result)
	})

	t.Run("handles unicode characters verification", func(t *testing.T) {
		password := "密码测试123🔒"
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)

		result := VerifyPassword(password, hashedPassword)

		assert.True(t, result)
	})

	t.Run("rejects invalid hash format", func(t *testing.T) {
		password := "testpassword"
		invalidHash := "invalid_hash_format"

		result := VerifyPassword(password, invalidHash)

		assert.False(t, result)
	})

	t.Run("case sensitive verification", func(t *testing.T) {
		password := "TestPassword"
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)

		upperResult := VerifyPassword("TESTPASSWORD", hashedPassword)
		lowerResult := VerifyPassword("testpassword", hashedPassword)
		correctResult := VerifyPassword("TestPassword", hashedPassword)

		assert.False(t, upperResult)
		assert.False(t, lowerResult)
		assert.True(t, correctResult)
	})

	t.Run("cross verification between different hashes", func(t *testing.T) {
		password1 := "password1"
		password2 := "password2"

		hash1, err1 := HashPassword(password1)
		hash2, err2 := HashPassword(password2)
		require.NoError(t, err1)
		require.NoError(t, err2)

		// Each password should only verify against its own hash
		assert.True(t, VerifyPassword(password1, hash1))
		assert.True(t, VerifyPassword(password2, hash2))
		assert.False(t, VerifyPassword(password1, hash2))
		assert.False(t, VerifyPassword(password2, hash1))
	})
}

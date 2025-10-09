package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilxqx/vef-framework-go/constants"
)

func TestPrefixKeyBuilder(t *testing.T) {
	t.Run("BuildWithoutPrefix", func(t *testing.T) {
		builder := NewPrefixKeyBuilder(constants.Empty)

		result := builder.Build("user", "123")
		assert.Equal(t, "user:123", result)
	})

	t.Run("BuildWithPrefix", func(t *testing.T) {
		builder := NewPrefixKeyBuilder("app")

		result := builder.Build("user", "123")
		assert.Equal(t, "app:user:123", result)
	})

	t.Run("BuildWithEmptyKeyParts", func(t *testing.T) {
		builder := NewPrefixKeyBuilder("app")

		// Should return prefix only, without trailing separator
		result := builder.Build()
		assert.Equal(t, "app", result)
	})

	t.Run("BuildWithEmptyKeyPartsNoPrefix", func(t *testing.T) {
		builder := NewPrefixKeyBuilder(constants.Empty)

		result := builder.Build()
		assert.Equal(t, constants.Empty, result)
	})

	t.Run("BuildWithSingleKeyPart", func(t *testing.T) {
		builder := NewPrefixKeyBuilder("app")

		result := builder.Build("user")
		assert.Equal(t, "app:user", result)
	})

	t.Run("BuildWithMultipleKeyParts", func(t *testing.T) {
		builder := NewPrefixKeyBuilder("app")

		result := builder.Build("user", "123", "profile")
		assert.Equal(t, "app:user:123:profile", result)
	})

	t.Run("BuildWithCustomSeparator", func(t *testing.T) {
		builder := NewPrefixKeyBuilderWithSeparator("app", "/")

		result := builder.Build("user", "123")
		assert.Equal(t, "app/user/123", result)
	})

	t.Run("BuildWithCustomSeparatorEmptyKeyParts", func(t *testing.T) {
		builder := NewPrefixKeyBuilderWithSeparator("app", "/")

		result := builder.Build()
		assert.Equal(t, "app", result)
	})

	t.Run("KeyHelperFunction", func(t *testing.T) {
		result := Key("user", "123")
		assert.Equal(t, "user:123", result)
	})

	t.Run("KeyHelperFunctionEmptyParts", func(t *testing.T) {
		result := Key()
		assert.Equal(t, constants.Empty, result)
	})
}

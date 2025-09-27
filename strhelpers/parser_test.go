package strhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTagAttrs(t *testing.T) {
	t.Run("single attribute without key", func(t *testing.T) {
		tag := "required"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 1)
		assert.Equal(t, "required", result["default"])
	})

	t.Run("single attribute with key", func(t *testing.T) {
		tag := "min=10"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 1)
		assert.Equal(t, "10", result["min"])
	})

	t.Run("multiple attributes", func(t *testing.T) {
		tag := "required,min=5,max=100"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 3)
		assert.Equal(t, "required", result["default"])
		assert.Equal(t, "5", result["min"])
		assert.Equal(t, "100", result["max"])
	})

	t.Run("attributes with spaces", func(t *testing.T) {
		tag := " required , min=5 , max=100 "

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 3)
		assert.Equal(t, "required", result["default"])
		assert.Equal(t, "5", result["min"])
		assert.Equal(t, "100", result["max"])
	})

	t.Run("empty tag", func(t *testing.T) {
		tag := ""

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 0)
	})

	t.Run("tag with empty attributes", func(t *testing.T) {
		tag := "required,,min=5"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 2)
		assert.Equal(t, "required", result["default"])
		assert.Equal(t, "5", result["min"])
	})

	t.Run("duplicate default attributes", func(t *testing.T) {
		tag := "required,optional"

		result := ParseTagAttrs(tag)

		// Should only keep the first default attribute
		assert.Len(t, result, 1)
		assert.Equal(t, "required", result["default"])
	})

	t.Run("attribute with empty value", func(t *testing.T) {
		tag := "min=,max=100"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 2)
		assert.Equal(t, "", result["min"])
		assert.Equal(t, "100", result["max"])
	})

	t.Run("complex tag", func(t *testing.T) {
		tag := "required,min=1,max=255,pattern=^[a-zA-Z0-9]+$"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 4)
		assert.Equal(t, "required", result["default"])
		assert.Equal(t, "1", result["min"])
		assert.Equal(t, "255", result["max"])
		assert.Equal(t, "^[a-zA-Z0-9]+$", result["pattern"])
	})
}

func TestParseTagArgs(t *testing.T) {
	t.Run("single argument without key", func(t *testing.T) {
		args := "search"

		result := ParseTagArgs(args)

		assert.Len(t, result, 1)
		assert.Equal(t, "search", result["default"])
	})

	t.Run("single argument with key", func(t *testing.T) {
		args := "q:golang"

		result := ParseTagArgs(args)

		assert.Len(t, result, 1)
		assert.Equal(t, "golang", result["q"])
	})

	t.Run("multiple arguments", func(t *testing.T) {
		args := "q:golang page:1 limit:10"

		result := ParseTagArgs(args)

		assert.Len(t, result, 3)
		assert.Equal(t, "golang", result["q"])
		assert.Equal(t, "1", result["page"])
		assert.Equal(t, "10", result["limit"])
	})

	t.Run("arguments with spaces", func(t *testing.T) {
		args := " q:golang page:1 limit:10 "

		result := ParseTagArgs(args)

		assert.Len(t, result, 3)
		assert.Equal(t, "golang", result["q"])
		assert.Equal(t, "1", result["page"])
		assert.Equal(t, "10", result["limit"])
	})

	t.Run("empty args", func(t *testing.T) {
		args := ""

		result := ParseTagArgs(args)

		assert.Len(t, result, 0)
	})

	t.Run("args with mixed separators", func(t *testing.T) {
		args := "q:golang page:1"

		result := ParseTagArgs(args)

		assert.Len(t, result, 2)
		assert.Equal(t, "golang", result["q"])
		assert.Equal(t, "1", result["page"])
	})

	t.Run("duplicate default arguments", func(t *testing.T) {
		args := "search filter"

		result := ParseTagArgs(args)

		// Should only keep the first default argument
		assert.Len(t, result, 1)
		assert.Equal(t, "search", result["default"])
	})

	t.Run("args with empty value", func(t *testing.T) {
		args := "q: page:1"

		result := ParseTagArgs(args)

		assert.Len(t, result, 2)
		assert.Equal(t, "", result["q"])
		assert.Equal(t, "1", result["page"])
	})

	t.Run("complex args", func(t *testing.T) {
		args := "q:web+framework category:backend sort:popularity order:desc"

		result := ParseTagArgs(args)

		assert.Len(t, result, 4)
		assert.Equal(t, "web+framework", result["q"])
		assert.Equal(t, "backend", result["category"])
		assert.Equal(t, "popularity", result["sort"])
		assert.Equal(t, "desc", result["order"])
	})

	t.Run("args with encoded characters", func(t *testing.T) {
		args := "q:hello%20world filter:type%3Darticle"

		result := ParseTagArgs(args)

		assert.Len(t, result, 2)
		assert.Equal(t, "hello%20world", result["q"])
		assert.Equal(t, "type%3Darticle", result["filter"])
	})
}

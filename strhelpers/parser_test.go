package strhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTagAttrs(t *testing.T) {
	t.Run("SingleAttributeWithoutKey", func(t *testing.T) {
		tag := "required"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 1, "Should parse one attribute")
		assert.Equal(t, "required", result[TagAttrDefaultKey], "Should set default key")
	})

	t.Run("SingleAttributeWithKey", func(t *testing.T) {
		tag := "min=10"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 1, "Should parse one attribute")
		assert.Equal(t, "10", result["min"], "Should parse key-value pair")
	})

	t.Run("MultipleAttributes", func(t *testing.T) {
		tag := "required,min=5,max=100"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 3, "Should parse three attributes")
		assert.Equal(t, "required", result[TagAttrDefaultKey], "Should set default key")
		assert.Equal(t, "5", result["min"], "Should parse min attribute")
		assert.Equal(t, "100", result["max"], "Should parse max attribute")
	})

	t.Run("AttributesWithSpaces", func(t *testing.T) {
		tag := " required , min=5 , max=100 "

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 3, "Should parse three attributes")
		assert.Equal(t, "required", result[TagAttrDefaultKey], "Should trim spaces and set default key")
		assert.Equal(t, "5", result["min"], "Should trim spaces and parse min attribute")
		assert.Equal(t, "100", result["max"], "Should trim spaces and parse max attribute")
	})

	t.Run("EmptyTag", func(t *testing.T) {
		tag := ""

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 0, "Should return empty map for empty tag")
	})

	t.Run("TagWithEmptyAttributes", func(t *testing.T) {
		tag := "required,,min=5"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 2, "Should skip empty attributes")
		assert.Equal(t, "required", result[TagAttrDefaultKey], "Should parse required attribute")
		assert.Equal(t, "5", result["min"], "Should parse min attribute")
	})

	t.Run("DuplicateDefaultAttributes", func(t *testing.T) {
		tag := "required,optional"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 1, "Should only keep first default attribute")
		assert.Equal(t, "required", result[TagAttrDefaultKey], "Should keep first default attribute")
	})

	t.Run("AttributeWithEmptyValue", func(t *testing.T) {
		tag := "min=,max=100"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 2, "Should parse both attributes")
		assert.Equal(t, "", result["min"], "Should allow empty value")
		assert.Equal(t, "100", result["max"], "Should parse max attribute")
	})

	t.Run("ComplexTag", func(t *testing.T) {
		tag := "required,min=1,max=255,pattern=^[a-zA-Z0-9]+$"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 4, "Should parse four attributes")
		assert.Equal(t, "required", result[TagAttrDefaultKey], "Should parse required attribute")
		assert.Equal(t, "1", result["min"], "Should parse min attribute")
		assert.Equal(t, "255", result["max"], "Should parse max attribute")
		assert.Equal(t, "^[a-zA-Z0-9]+$", result["pattern"], "Should parse pattern attribute")
	})
}

func TestParseTagArgs(t *testing.T) {
	t.Run("SingleArgumentWithoutKey", func(t *testing.T) {
		args := "search"

		result := ParseTagArgs(args)

		assert.Len(t, result, 1, "Should parse one argument")
		assert.Equal(t, "search", result[TagAttrDefaultKey], "Should set default key")
	})

	t.Run("SingleArgumentWithKey", func(t *testing.T) {
		args := "q:golang"

		result := ParseTagArgs(args)

		assert.Len(t, result, 1, "Should parse one argument")
		assert.Equal(t, "golang", result["q"], "Should parse key-value pair")
	})

	t.Run("MultipleArguments", func(t *testing.T) {
		args := "q:golang page:1 limit:10"

		result := ParseTagArgs(args)

		assert.Len(t, result, 3, "Should parse three arguments")
		assert.Equal(t, "golang", result["q"], "Should parse q argument")
		assert.Equal(t, "1", result["page"], "Should parse page argument")
		assert.Equal(t, "10", result["limit"], "Should parse limit argument")
	})

	t.Run("ArgumentsWithSpaces", func(t *testing.T) {
		args := " q:golang page:1 limit:10 "

		result := ParseTagArgs(args)

		assert.Len(t, result, 3, "Should parse three arguments")
		assert.Equal(t, "golang", result["q"], "Should trim spaces and parse q argument")
		assert.Equal(t, "1", result["page"], "Should trim spaces and parse page argument")
		assert.Equal(t, "10", result["limit"], "Should trim spaces and parse limit argument")
	})

	t.Run("EmptyArgs", func(t *testing.T) {
		args := ""

		result := ParseTagArgs(args)

		assert.Len(t, result, 0, "Should return empty map for empty args")
	})

	t.Run("ArgsWithMixedSeparators", func(t *testing.T) {
		args := "q:golang page:1"

		result := ParseTagArgs(args)

		assert.Len(t, result, 2, "Should parse two arguments")
		assert.Equal(t, "golang", result["q"], "Should parse q argument")
		assert.Equal(t, "1", result["page"], "Should parse page argument")
	})

	t.Run("DuplicateDefaultArguments", func(t *testing.T) {
		args := "search filter"

		result := ParseTagArgs(args)

		assert.Len(t, result, 1, "Should only keep first default argument")
		assert.Equal(t, "search", result[TagAttrDefaultKey], "Should keep first default argument")
	})

	t.Run("ArgsWithEmptyValue", func(t *testing.T) {
		args := "q: page:1"

		result := ParseTagArgs(args)

		assert.Len(t, result, 2, "Should parse both arguments")
		assert.Equal(t, "", result["q"], "Should allow empty value")
		assert.Equal(t, "1", result["page"], "Should parse page argument")
	})

	t.Run("ComplexArgs", func(t *testing.T) {
		args := "q:web+framework category:backend sort:popularity order:desc"

		result := ParseTagArgs(args)

		assert.Len(t, result, 4, "Should parse four arguments")
		assert.Equal(t, "web+framework", result["q"], "Should parse q argument with special chars")
		assert.Equal(t, "backend", result["category"], "Should parse category argument")
		assert.Equal(t, "popularity", result["sort"], "Should parse sort argument")
		assert.Equal(t, "desc", result["order"], "Should parse order argument")
	})

	t.Run("ArgsWithEncodedCharacters", func(t *testing.T) {
		args := "q:hello%20world filter:type%3Darticle"

		result := ParseTagArgs(args)

		assert.Len(t, result, 2, "Should parse two arguments")
		assert.Equal(t, "hello%20world", result["q"], "Should preserve encoded characters")
		assert.Equal(t, "type%3Darticle", result["filter"], "Should preserve encoded characters")
	})
}

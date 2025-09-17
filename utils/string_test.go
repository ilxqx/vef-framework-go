package utils

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

		// Empty string is now skipped, so no entries
		assert.Len(t, result, 0)
	})

	t.Run("tag with empty attribute", func(t *testing.T) {
		tag := "required,,max=100"

		result := ParseTagAttrs(tag)

		// Empty attributes are now skipped
		assert.Len(t, result, 2)
		assert.Equal(t, "required", result["default"])
		assert.Equal(t, "100", result["max"])
	})

	t.Run("duplicate default attributes", func(t *testing.T) {
		tag := "required,validate"

		result := ParseTagAttrs(tag)

		// Should only keep the first default attribute
		assert.Len(t, result, 1)
		assert.Equal(t, "required", result["default"])
	})

	t.Run("attribute with empty value", func(t *testing.T) {
		tag := "required,min="

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 2)
		assert.Equal(t, "required", result["default"])
		assert.Equal(t, "", result["min"])
	})

	t.Run("complex attribute values", func(t *testing.T) {
		tag := "regex=^[a-zA-Z0-9]+$,format=email"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 2)
		assert.Equal(t, "^[a-zA-Z0-9]+$", result["regex"])
		assert.Equal(t, "email", result["format"])
	})

	t.Run("attribute with multiple equals", func(t *testing.T) {
		tag := "url=http://example.com:8080/path"

		result := ParseTagAttrs(tag)

		assert.Len(t, result, 1)
		assert.Equal(t, "http://example.com:8080/path", result["url"])
	})
}

func TestParseQueryString(t *testing.T) {
	t.Run("single parameter without key", func(t *testing.T) {
		query := "search"

		result := ParseQueryString(query)

		assert.Len(t, result, 1)
		assert.Equal(t, "search", result["default"])
	})

	t.Run("single parameter with key", func(t *testing.T) {
		query := "name=john"

		result := ParseQueryString(query)

		assert.Len(t, result, 1)
		assert.Equal(t, "john", result["name"])
	})

	t.Run("multiple parameters", func(t *testing.T) {
		query := "name=john&age=30&city=newyork"

		result := ParseQueryString(query)

		assert.Len(t, result, 3)
		assert.Equal(t, "john", result["name"])
		assert.Equal(t, "30", result["age"])
		assert.Equal(t, "newyork", result["city"])
	})

	t.Run("parameters with spaces", func(t *testing.T) {
		query := " name=john & age=30 & city=newyork "

		result := ParseQueryString(query)

		assert.Len(t, result, 3)
		assert.Equal(t, "john", result["name"])
		assert.Equal(t, "30", result["age"])
		assert.Equal(t, "newyork", result["city"])
	})

	t.Run("empty query", func(t *testing.T) {
		query := ""

		result := ParseQueryString(query)

		// Empty string is now skipped, so no entries
		assert.Len(t, result, 0)
	})

	t.Run("query with empty parameter", func(t *testing.T) {
		query := "name=john&&age=30"

		result := ParseQueryString(query)

		// Empty parameters are now skipped
		assert.Len(t, result, 2)
		assert.Equal(t, "john", result["name"])
		assert.Equal(t, "30", result["age"])
	})

	t.Run("duplicate default parameters", func(t *testing.T) {
		query := "search&filter"

		result := ParseQueryString(query)

		// Should only keep the first default parameter
		assert.Len(t, result, 1)
		assert.Equal(t, "search", result["default"])
	})

	t.Run("parameter with empty value", func(t *testing.T) {
		query := "name=john&search="

		result := ParseQueryString(query)

		assert.Len(t, result, 2)
		assert.Equal(t, "john", result["name"])
		assert.Equal(t, "", result["search"])
	})

	t.Run("complex parameter values", func(t *testing.T) {
		query := "url=http://example.com&email=user@example.com"

		result := ParseQueryString(query)

		assert.Len(t, result, 2)
		assert.Equal(t, "http://example.com", result["url"])
		assert.Equal(t, "user@example.com", result["email"])
	})

	t.Run("parameter with multiple equals", func(t *testing.T) {
		query := "formula=a=b+c"

		result := ParseQueryString(query)

		assert.Len(t, result, 1)
		assert.Equal(t, "a=b+c", result["formula"])
	})

	t.Run("URL encoded values", func(t *testing.T) {
		query := "message=hello%20world&space=%20"

		result := ParseQueryString(query)

		assert.Len(t, result, 2)
		// Note: This function doesn't URL decode, it just parses
		assert.Equal(t, "hello%20world", result["message"])
		assert.Equal(t, "%20", result["space"])
	})
}

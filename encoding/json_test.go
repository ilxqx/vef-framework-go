package encoding

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStruct for encoding/decoding tests.
type TestStruct struct {
	Name    string    `json:"name"`
	Age     int       `json:"age"`
	Email   string    `json:"email,omitempty"`
	Active  bool      `json:"active"`
	Score   float64   `json:"score"`
	Created time.Time `json:"created"`
}

func TestToJSON(t *testing.T) {
	t.Run("Valid struct", func(t *testing.T) {
		input := TestStruct{
			Name:   "John Doe",
			Age:    30,
			Active: true,
		}

		result, err := ToJSON(input)
		require.NoError(t, err)
		assert.Contains(t, result, "\"name\":\"John Doe\"")
		assert.Contains(t, result, "\"age\":30")
		assert.Contains(t, result, "\"active\":true")
	})

	t.Run("Nil input", func(t *testing.T) {
		result, err := ToJSON(nil)
		require.NoError(t, err)
		assert.Equal(t, "null", result)
	})

	t.Run("Empty struct", func(t *testing.T) {
		input := TestStruct{}

		result, err := ToJSON(input)
		require.NoError(t, err)
		assert.Contains(t, result, "\"name\":\"\"")
		assert.Contains(t, result, "\"age\":0")
		assert.Contains(t, result, "\"active\":false")
	})
}

func TestFromJSON(t *testing.T) {
	t.Run("Valid JSON", func(t *testing.T) {
		input := `{"name":"John Doe","age":30,"email":"john@example.com","active":true,"score":95.5}`

		result, err := FromJSON[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Equal(t, "john@example.com", result.Email)
		assert.Equal(t, true, result.Active)
		assert.Equal(t, 95.5, result.Score)
	})

	t.Run("Partial JSON", func(t *testing.T) {
		input := `{"name":"Jane Doe"}`

		result, err := FromJSON[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Jane Doe", result.Name)
		assert.Equal(t, 0, result.Age)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		input := `{"name":"John Doe","age":}`

		result, err := FromJSON[TestStruct](input)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Empty JSON", func(t *testing.T) {
		input := `{}`

		result, err := FromJSON[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

func TestDecodeJSON(t *testing.T) {
	t.Run("Decode into struct pointer", func(t *testing.T) {
		input := `{"name":"John Doe","age":30,"active":true}`

		var result TestStruct

		err := DecodeJSON(input, &result)
		require.NoError(t, err)

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.True(t, result.Active)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		input := `{"name":"John Doe","age":}`

		var result TestStruct

		err := DecodeJSON(input, &result)
		assert.Error(t, err)
	})
}

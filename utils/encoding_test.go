package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test struct for encoding/decoding tests
type TestStruct struct {
	Name    string    `json:"name"`
	Age     int       `json:"age"`
	Email   string    `json:"email,omitempty"`
	Active  bool      `json:"active"`
	Score   float64   `json:"score"`
	Created time.Time `json:"created"`
}

func TestToJSON(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
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

	t.Run("nil input", func(t *testing.T) {
		result, err := ToJSON(nil)
		require.NoError(t, err)
		assert.Equal(t, "null", result)
	})

	t.Run("empty struct", func(t *testing.T) {
		input := TestStruct{}

		result, err := ToJSON(input)
		require.NoError(t, err)
		assert.Contains(t, result, "\"name\":\"\"")
		assert.Contains(t, result, "\"age\":0")
		assert.Contains(t, result, "\"active\":false")
	})
}

func TestFromJSON(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
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

	t.Run("partial JSON", func(t *testing.T) {
		input := `{"name":"Jane Doe"}`

		result, err := FromJSON[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Jane Doe", result.Name)
		assert.Equal(t, 0, result.Age)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		input := `{"name":"John Doe","age":}`

		result, err := FromJSON[TestStruct](input)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty JSON", func(t *testing.T) {
		input := `{}`

		result, err := FromJSON[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

func TestToXML(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		input := TestStruct{
			Name:   "John Doe",
			Age:    30,
			Active: true,
		}

		result, err := ToXML(input)
		require.NoError(t, err)
		assert.Contains(t, result, "<TestStruct>")
		assert.Contains(t, result, "<Name>John Doe</Name>")
		assert.Contains(t, result, "<Age>30</Age>")
		assert.Contains(t, result, "<Active>true</Active>")
	})

	t.Run("empty struct", func(t *testing.T) {
		input := TestStruct{}

		result, err := ToXML(input)
		require.NoError(t, err)
		assert.Contains(t, result, "<TestStruct>")
		assert.Contains(t, result, "<Name></Name>")
		assert.Contains(t, result, "<Age>0</Age>")
		assert.Contains(t, result, "<Active>false</Active>")
	})
}

func TestFromXML(t *testing.T) {
	t.Run("valid XML", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</Age><Active>true</Active><Score>95.5</Score></TestStruct>`

		result, err := FromXML[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Equal(t, true, result.Active)
		assert.Equal(t, 95.5, result.Score)
	})

	t.Run("partial XML", func(t *testing.T) {
		input := `<TestStruct><Name>Jane Doe</Name></TestStruct>`

		result, err := FromXML[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Jane Doe", result.Name)
		assert.Equal(t, 0, result.Age)
	})

	t.Run("invalid XML", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</TestStruct>`

		result, err := FromXML[TestStruct](input)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty XML", func(t *testing.T) {
		input := `<TestStruct></TestStruct>`

		result, err := FromXML[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

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

func TestToMap(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		input := TestStruct{
			Name:    "John Doe",
			Age:     30,
			Email:   "john@example.com",
			Active:  true,
			Score:   95.5,
			Created: testTime,
		}

		result, err := ToMap(input)
		require.NoError(t, err)

		assert.Equal(t, "John Doe", result["name"])
		assert.Equal(t, 30, result["age"])
		assert.Equal(t, "john@example.com", result["email"])
		assert.Equal(t, true, result["active"])
		assert.Equal(t, 95.5, result["score"])
		// Time field is converted to a map by mapstructure, so we check if it exists
		assert.Contains(t, result, "created")
	})

	t.Run("pointer to struct", func(t *testing.T) {
		input := &TestStruct{
			Name: "Jane Doe",
			Age:  25,
		}

		result, err := ToMap(input)
		require.NoError(t, err)

		assert.Equal(t, "Jane Doe", result["name"])
		assert.Equal(t, 25, result["age"])
	})

	t.Run("non-struct value", func(t *testing.T) {
		input := "not a struct"

		result, err := ToMap(input)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "must be a struct")
	})

	t.Run("slice input", func(t *testing.T) {
		input := []int{1, 2, 3}

		result, err := ToMap(input)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestFromMap(t *testing.T) {
	t.Run("valid map to struct", func(t *testing.T) {
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		input := map[string]any{
			"name":    "John Doe",
			"age":     30,
			"email":   "john@example.com",
			"active":  true,
			"score":   95.5,
			"created": testTime,
		}

		result, err := FromMap[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Equal(t, "john@example.com", result.Email)
		assert.Equal(t, true, result.Active)
		assert.Equal(t, 95.5, result.Score)
		assert.Equal(t, testTime, result.Created)
	})

	t.Run("partial map", func(t *testing.T) {
		input := map[string]any{
			"name": "Jane Doe",
			"age":  25,
		}

		result, err := FromMap[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Jane Doe", result.Name)
		assert.Equal(t, 25, result.Age)
		assert.Equal(t, "", result.Email)
		assert.Equal(t, false, result.Active)
	})

	t.Run("empty map", func(t *testing.T) {
		input := map[string]any{}

		result, err := FromMap[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

func TestNewMapDecoder(t *testing.T) {
	t.Run("create decoder for struct", func(t *testing.T) {
		var result TestStruct
		decoder, err := NewMapDecoder(&result)
		require.NoError(t, err)
		assert.NotNil(t, decoder)
	})

	t.Run("decode with created decoder", func(t *testing.T) {
		var result TestStruct
		decoder, err := NewMapDecoder(&result)
		require.NoError(t, err)

		input := map[string]any{
			"name": "Test User",
			"age":  35,
		}

		err = decoder.Decode(input)
		require.NoError(t, err)

		assert.Equal(t, "Test User", result.Name)
		assert.Equal(t, 35, result.Age)
	})
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

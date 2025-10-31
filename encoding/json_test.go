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

func TestToJson(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
	}{
		{
			name: "ValidStruct",
			input: TestStruct{
				Name:   "John Doe",
				Age:    30,
				Active: true,
			},
			expected: []string{`"name":"John Doe"`, `"age":30`, `"active":true`},
		},
		{
			name:     "NilInput",
			input:    nil,
			expected: []string{"null"},
		},
		{
			name:     "EmptyStruct",
			input:    TestStruct{},
			expected: []string{`"name":""`, `"age":0`, `"active":false`},
		},
		{
			name: "StructWithAllFields",
			input: TestStruct{
				Name:    "Jane Doe",
				Age:     25,
				Email:   "jane@example.com",
				Active:  true,
				Score:   95.5,
				Created: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: []string{`"name":"Jane Doe"`, `"email":"jane@example.com"`, `"score":95.5`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToJson(tt.input)
			require.NoError(t, err)

			for _, exp := range tt.expected {
				assert.Contains(t, result, exp)
			}
		})
	}
}

func TestFromJson(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		input := `{"name":"John Doe","age":30,"email":"john@example.com","active":true,"score":95.5}`
		result, err := FromJson[TestStruct](input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Equal(t, "john@example.com", result.Email)
		assert.True(t, result.Active)
		assert.Equal(t, 95.5, result.Score)
	})

	t.Run("PartialJSON", func(t *testing.T) {
		input := `{"name":"Jane Doe"}`
		result, err := FromJson[TestStruct](input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Jane Doe", result.Name)
		assert.Equal(t, 0, result.Age)
		assert.False(t, result.Active)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		input := `{"name":"John Doe","age":}`
		_, err := FromJson[TestStruct](input)
		assert.Error(t, err)
	})

	t.Run("EmptyJSON", func(t *testing.T) {
		input := `{}`
		result, err := FromJson[TestStruct](input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
		assert.False(t, result.Active)
	})

	t.Run("JSONWithExtraFields", func(t *testing.T) {
		input := `{"name":"John Doe","age":30,"extra":"field"}`
		result, err := FromJson[TestStruct](input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
	})
}

func TestDecodeJson(t *testing.T) {
	t.Run("DecodeIntoStructPointer", func(t *testing.T) {
		input := `{"name":"John Doe","age":30,"active":true}`

		var result TestStruct

		err := DecodeJson(input, &result)
		require.NoError(t, err)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.True(t, result.Active)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		input := `{"name":"John Doe","age":}`

		var result TestStruct

		err := DecodeJson(input, &result)
		assert.Error(t, err)
	})

	t.Run("EmptyJSON", func(t *testing.T) {
		input := `{}`

		var result TestStruct

		err := DecodeJson(input, &result)
		require.NoError(t, err)
		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

func TestJsonRoundTrip(t *testing.T) {
	created := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	input := TestStruct{
		Name:    "Jane Doe",
		Age:     25,
		Email:   "jane@example.com",
		Active:  true,
		Score:   88.5,
		Created: created,
	}

	encoded, err := ToJson(input)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded, err := FromJson[TestStruct](encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)
	assert.Equal(t, input.Name, decoded.Name)
	assert.Equal(t, input.Age, decoded.Age)
	assert.Equal(t, input.Email, decoded.Email)
	assert.Equal(t, input.Active, decoded.Active)
	assert.Equal(t, input.Score, decoded.Score)
	assert.Equal(t, input.Created.Unix(), decoded.Created.Unix())
}

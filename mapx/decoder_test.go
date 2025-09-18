package mapx

import (
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test struct for encoding/decoding tests
type TestStruct struct {
	Name       string        `json:"name"`
	Age        int           `json:"age"`
	Email      string        `json:"email,omitempty"`
	Active     bool          `json:"active"`
	Score      float64       `json:"score"`
	Created    time.Time     `json:"created"`
	Duration   time.Duration `json:"duration"`
	Website    *url.URL      `json:"website"`
	IP         net.IP        `json:"ip"`
	Unexported string        // This field should be ignored by default
}

type EmbeddedStruct struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type StructWithEmbedding struct {
	Name     string         `json:"name"`
	Embedded EmbeddedStruct `json:"embedded,inline"`
}

func TestNewDecoder(t *testing.T) {
	t.Run("create decoder with default options", func(t *testing.T) {
		var result TestStruct
		decoder, err := NewDecoder(&result)
		require.NoError(t, err)
		assert.NotNil(t, decoder)
	})

	t.Run("create decoder with custom options", func(t *testing.T) {
		var result TestStruct
		decoder, err := NewDecoder(&result, WithTagName("custom"), WithErrorUnused())
		require.NoError(t, err)
		assert.NotNil(t, decoder)
	})
}

func TestToMap(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		testURL, _ := url.Parse("https://example.com")
		input := TestStruct{
			Name:     "John Doe",
			Age:      30,
			Email:    "john@example.com",
			Active:   true,
			Score:    95.5,
			Created:  testTime,
			Duration: time.Hour,
			Website:  testURL,
			IP:       net.ParseIP("192.168.1.1"),
		}

		result, err := ToMap(input)
		require.NoError(t, err)

		assert.Equal(t, "John Doe", result["name"])
		assert.Equal(t, 30, result["age"])
		assert.Equal(t, "john@example.com", result["email"])
		assert.Equal(t, true, result["active"])
		assert.Equal(t, 95.5, result["score"])
		assert.Contains(t, result, "created")
		assert.Contains(t, result, "duration")
		assert.Contains(t, result, "website")
		assert.Contains(t, result, "ip")
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

	t.Run("struct with embedding", func(t *testing.T) {
		input := StructWithEmbedding{
			Name: "Test",
			Embedded: EmbeddedStruct{
				ID:   123,
				Type: "example",
			},
		}

		result, err := ToMap(input)
		require.NoError(t, err)

		assert.Equal(t, "Test", result["name"])
		// Check if embedded fields are inlined
		assert.Equal(t, 123, result["id"])
		assert.Equal(t, "example", result["type"])
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
		assert.Contains(t, err.Error(), "must be a struct")
	})

	t.Run("with custom tag name", func(t *testing.T) {
		type CustomTagStruct struct {
			Name string `custom:"full_name"`
			Age  int    `custom:"years"`
		}

		input := CustomTagStruct{Name: "John", Age: 30}
		result, err := ToMap(input, WithTagName("custom"))
		require.NoError(t, err)

		assert.Equal(t, "John", result["full_name"])
		assert.Equal(t, 30, result["years"])
	})
}

func TestFromMap(t *testing.T) {
	t.Run("valid map to struct", func(t *testing.T) {
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		input := map[string]any{
			"name":     "John Doe",
			"age":      30,
			"email":    "john@example.com",
			"active":   true,
			"score":    95.5,
			"created":  testTime,
			"duration": "1h",
			"website":  "https://example.com",
			"ip":       "192.168.1.1",
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
		assert.Equal(t, time.Hour, result.Duration)
		assert.Equal(t, "https://example.com", result.Website.String())
		assert.Equal(t, "192.168.1.1", result.IP.String())
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

	t.Run("map with embedding", func(t *testing.T) {
		input := map[string]any{
			"name": "Test",
			"id":   123,
			"type": "example",
		}

		result, err := FromMap[StructWithEmbedding](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Test", result.Name)
		assert.Equal(t, 123, result.Embedded.ID)
		assert.Equal(t, "example", result.Embedded.Type)
	})

	t.Run("non-struct type parameter", func(t *testing.T) {
		input := map[string]any{"value": "test"}

		result, err := FromMap[string](input)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "must be a struct")
	})

	t.Run("with custom tag name", func(t *testing.T) {
		type CustomTagStruct struct {
			Name string `custom:"full_name"`
			Age  int    `custom:"years"`
		}

		input := map[string]any{
			"full_name": "John",
			"years":     30,
		}

		result, err := FromMap[CustomTagStruct](input, WithTagName("custom"))
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
	})
}

func TestDecoderOptions(t *testing.T) {
	t.Run("WithTagName", func(t *testing.T) {
		type TestStruct struct {
			Name string `yaml:"full_name"`
		}

		input := map[string]any{"full_name": "John"}
		result, err := FromMap[TestStruct](input, WithTagName("yaml"))
		require.NoError(t, err)
		assert.Equal(t, "John", result.Name)
	})

	t.Run("WithIgnoreUntaggedFields", func(t *testing.T) {
		type TestStruct struct {
			Name          string `json:"name"`
			UntaggedField string
		}

		input := map[string]any{
			"name":          "John",
			"UntaggedField": "should be ignored",
		}

		result, err := FromMap[TestStruct](input, WithIgnoreUntaggedFields(true))
		require.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, "", result.UntaggedField) // Should be empty because it's ignored
	})

	t.Run("WithWeaklyTypedInput", func(t *testing.T) {
		type TestStruct struct {
			Age int `json:"age"`
		}

		// String instead of int
		input := map[string]any{"age": "30"}
		result, err := FromMap[TestStruct](input, WithWeaklyTypedInput())
		require.NoError(t, err)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("WithZeroFields", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		// Create a decoder with ZeroFields option
		input := map[string]any{"name": "New Name"}
		result, err := FromMap[TestStruct](input, WithZeroFields())
		require.NoError(t, err)

		assert.Equal(t, "New Name", result.Name)
		assert.Equal(t, 0, result.Age) // Should be zero by default
	})

	t.Run("WithMetadata", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		var metadata Metadata
		input := map[string]any{
			"name":  "John",
			"age":   30,
			"extra": "unused field",
		}

		result, err := FromMap[TestStruct](input, WithMetadata(&metadata))
		require.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Contains(t, metadata.Unused, "extra")
	})
}

func TestComplexTypeConversions(t *testing.T) {
	t.Run("time conversion", func(t *testing.T) {
		type TimeStruct struct {
			Created time.Time `json:"created"`
		}

		input := map[string]any{
			"created": "2023-01-01T12:00:00Z",
		}

		result, err := FromMap[TimeStruct](input)
		require.NoError(t, err)
		expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
		assert.Equal(t, expectedTime, result.Created)
	})

	t.Run("duration conversion", func(t *testing.T) {
		type DurationStruct struct {
			Timeout time.Duration `json:"timeout"`
		}

		input := map[string]any{
			"timeout": "5m30s",
		}

		result, err := FromMap[DurationStruct](input)
		require.NoError(t, err)
		expected, _ := time.ParseDuration("5m30s")
		assert.Equal(t, expected, result.Timeout)
	})

	t.Run("URL conversion", func(t *testing.T) {
		type URLStruct struct {
			Website *url.URL `json:"website"`
		}

		input := map[string]any{
			"website": "https://example.com/path?param=value",
		}

		result, err := FromMap[URLStruct](input)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com/path?param=value", result.Website.String())
	})

	t.Run("IP conversion", func(t *testing.T) {
		type IPStruct struct {
			Address net.IP `json:"address"`
		}

		input := map[string]any{
			"address": "192.168.1.100",
		}

		result, err := FromMap[IPStruct](input)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.100", result.Address.String())
	})
}

func TestRoundTripConversion(t *testing.T) {
	t.Run("struct to map and back", func(t *testing.T) {
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		original := TestStruct{
			Name:     "John Doe",
			Age:      30,
			Email:    "john@example.com",
			Active:   true,
			Score:    95.5,
			Created:  testTime,
			Duration: time.Hour * 2,
		}

		// Convert to map
		mapResult, err := ToMap(original)
		require.NoError(t, err)

		// Convert back to struct
		structResult, err := FromMap[TestStruct](mapResult)
		require.NoError(t, err)

		assert.Equal(t, original.Name, structResult.Name)
		assert.Equal(t, original.Age, structResult.Age)
		assert.Equal(t, original.Email, structResult.Email)
		assert.Equal(t, original.Active, structResult.Active)
		assert.Equal(t, original.Score, structResult.Score)
		// Note: Time and Duration might have slight differences due to encoding/decoding
	})
}

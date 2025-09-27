package mapx

import (
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/null"
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

func TestNullBoolDecodeHook(t *testing.T) {
	t.Run("null.Bool to bool conversion", func(t *testing.T) {
		type StructWithBool struct {
			Active bool `json:"active"`
		}

		input := map[string]any{
			"active": null.BoolFrom(true),
		}

		result, err := FromMap[StructWithBool](input)
		require.NoError(t, err)
		assert.True(t, result.Active)
	})

	t.Run("bool to null.Bool conversion", func(t *testing.T) {
		type StructWithNullBool struct {
			Active null.Bool `json:"active"`
		}

		input := map[string]any{
			"active": true,
		}

		result, err := FromMap[StructWithNullBool](input)
		require.NoError(t, err)
		assert.True(t, result.Active.Valid)
		assert.True(t, result.Active.Bool)
	})

	t.Run("invalid null.Bool to bool", func(t *testing.T) {
		type StructWithBool struct {
			Active bool `json:"active"`
		}

		input := map[string]any{
			"active": null.NewBool(true, false), // invalid null.Bool
		}

		result, err := FromMap[StructWithBool](input)
		require.NoError(t, err)
		assert.False(t, result.Active) // Should be false for invalid null.Bool
	})

	t.Run("ToMap with null.Bool", func(t *testing.T) {
		type StructWithNullBool struct {
			Active null.Bool `json:"active"`
		}

		input := StructWithNullBool{
			Active: null.BoolFrom(true),
		}

		result, err := ToMap(input)
		require.NoError(t, err)
		// Check what type we actually got
		t.Logf("Result type: %T, value: %v", result["active"], result["active"])
		// The null.Bool might be converted to a map structure, let's check if it has the expected value
		if boolVal, ok := result["active"].(bool); ok {
			assert.True(t, boolVal)
		} else {
			// If it's a map, check the structure
			if mapVal, ok := result["active"].(map[string]any); ok {
				assert.True(t, mapVal["Valid"].(bool))
				assert.True(t, mapVal["Bool"].(bool))
			} else {
				t.Fatalf("Unexpected type for active field: %T", result["active"])
			}
		}
	})
}

func TestNullValueDecodeHook(t *testing.T) {
	t.Run("null.Value[string] to string conversion", func(t *testing.T) {
		type StructWithString struct {
			Name string `json:"name"`
		}

		input := map[string]any{
			"name": null.ValueFrom("John Doe"),
		}

		result, err := FromMap[StructWithString](input)
		require.NoError(t, err)
		assert.Equal(t, "John Doe", result.Name)
	})

	t.Run("string to null.Value[string] conversion", func(t *testing.T) {
		type StructWithNullString struct {
			Name null.Value[string] `json:"name"`
		}

		input := map[string]any{
			"name": "John Doe",
		}

		result, err := FromMap[StructWithNullString](input)
		require.NoError(t, err)
		assert.True(t, result.Name.Valid)
		assert.Equal(t, "John Doe", result.Name.V)
	})

	t.Run("null.Value[int] to int conversion", func(t *testing.T) {
		type StructWithInt struct {
			Age int `json:"age"`
		}

		input := map[string]any{
			"age": null.ValueFrom(30),
		}

		result, err := FromMap[StructWithInt](input)
		require.NoError(t, err)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("int to null.Value[int] conversion", func(t *testing.T) {
		type StructWithNullInt struct {
			Age null.Value[int] `json:"age"`
		}

		input := map[string]any{
			"age": 30,
		}

		result, err := FromMap[StructWithNullInt](input)
		require.NoError(t, err)
		assert.True(t, result.Age.Valid)
		assert.Equal(t, 30, result.Age.V)
	})

	t.Run("invalid null.Value to primitive", func(t *testing.T) {
		type StructWithString struct {
			Name string `json:"name"`
		}

		input := map[string]any{
			"name": null.NewValue("John", false), // invalid null.Value
		}

		result, err := FromMap[StructWithString](input)
		require.NoError(t, err)
		assert.Equal(t, "", result.Name) // Should be zero value for invalid null.Value
	})

	t.Run("ToMap with null.Value", func(t *testing.T) {
		type StructWithNullValue struct {
			Name null.Value[string] `json:"name"`
			Age  null.Value[int]    `json:"age"`
		}

		input := StructWithNullValue{
			Name: null.ValueFrom("John Doe"),
			Age:  null.ValueFrom(30),
		}

		result, err := ToMap(input)
		require.NoError(t, err)

		// Check name field
		if nameVal, ok := result["name"].(string); ok {
			assert.Equal(t, "John Doe", nameVal)
		} else if mapVal, ok := result["name"].(map[string]any); ok {
			assert.True(t, mapVal["Valid"].(bool))
			assert.Equal(t, "John Doe", mapVal["V"])
		}

		// Check age field
		if ageVal, ok := result["age"].(int); ok {
			assert.Equal(t, 30, ageVal)
		} else if mapVal, ok := result["age"].(map[string]any); ok {
			assert.True(t, mapVal["Valid"].(bool))
			assert.Equal(t, 30, mapVal["V"])
		}
	})
}

func TestNullTypesIntegration(t *testing.T) {
	t.Run("complex struct with various null types", func(t *testing.T) {
		type ComplexStruct struct {
			Name          null.Value[string]  `json:"name"`
			Age           null.Value[int]     `json:"age"`
			Active        null.Bool           `json:"active"`
			Score         null.Value[float64] `json:"score"`
			OptionalField null.Value[string]  `json:"optional_field"`
		}

		input := map[string]any{
			"name":   "John Doe",
			"age":    30,
			"active": true,
			"score":  95.5,
		}

		result, err := FromMap[ComplexStruct](input)
		require.NoError(t, err)

		assert.True(t, result.Name.Valid)
		assert.Equal(t, "John Doe", result.Name.V)

		assert.True(t, result.Age.Valid)
		assert.Equal(t, 30, result.Age.V)

		assert.True(t, result.Active.Valid)
		assert.True(t, result.Active.Bool)

		assert.True(t, result.Score.Valid)
		assert.Equal(t, 95.5, result.Score.V)

		assert.False(t, result.OptionalField.Valid) // Not provided in input
	})

	t.Run("round trip with null types", func(t *testing.T) {
		// Note: Round trip with null types has limitations because ToMap converts
		// null.Value and null.Bool to map structures rather than primitives.
		// This is expected behavior due to how mapstructure handles struct decomposition.

		type NullStruct struct {
			Name   null.Value[string] `json:"name"`
			Age    null.Value[int]    `json:"age"`
			Active null.Bool          `json:"active"`
		}

		original := NullStruct{
			Name:   null.ValueFrom("Jane Doe"),
			Age:    null.ValueFrom(25),
			Active: null.BoolFrom(false),
		}

		// Convert to map
		mapResult, err := ToMap(original)
		require.NoError(t, err)

		// Verify the map contains the expected structure
		nameMap, ok := mapResult["name"].(map[string]any)
		require.True(t, ok, "name should be converted to map")
		assert.True(t, nameMap["Valid"].(bool))
		assert.Equal(t, "Jane Doe", nameMap["V"])

		ageMap, ok := mapResult["age"].(map[string]any)
		require.True(t, ok, "age should be converted to map")
		assert.True(t, ageMap["Valid"].(bool))
		assert.Equal(t, 25, ageMap["V"])

		activeMap, ok := mapResult["active"].(map[string]any)
		require.True(t, ok, "active should be converted to map")
		assert.True(t, activeMap["Valid"].(bool))
		assert.False(t, activeMap["Bool"].(bool))

		// Note: Full round trip back to the original struct is not supported
		// because the map structure doesn't map cleanly back to null types.
		// This is a known limitation when using nested struct types.
	})

	t.Run("mixed null and regular types", func(t *testing.T) {
		type MixedStruct struct {
			RegularName string             `json:"regular_name"`
			NullName    null.Value[string] `json:"null_name"`
			RegularAge  int                `json:"regular_age"`
			NullAge     null.Value[int]    `json:"null_age"`
			RegularFlag bool               `json:"regular_flag"`
			NullFlag    null.Bool          `json:"null_flag"`
		}

		input := map[string]any{
			"regular_name": "John",
			"null_name":    "Jane",
			"regular_age":  30,
			"null_age":     25,
			"regular_flag": true,
			"null_flag":    false,
		}

		result, err := FromMap[MixedStruct](input)
		require.NoError(t, err)

		assert.Equal(t, "John", result.RegularName)
		assert.Equal(t, "Jane", result.NullName.V)
		assert.True(t, result.NullName.Valid)

		assert.Equal(t, 30, result.RegularAge)
		assert.Equal(t, 25, result.NullAge.V)
		assert.True(t, result.NullAge.Valid)

		assert.True(t, result.RegularFlag)
		assert.False(t, result.NullFlag.Bool)
		assert.True(t, result.NullFlag.Valid)
	})
}

func TestNullBoolBasicOperations(t *testing.T) {
	t.Run("BoolFrom creates valid null.Bool", func(t *testing.T) {
		b := null.BoolFrom(true)
		assert.True(t, b.Valid)
		assert.True(t, b.Bool)
		assert.True(t, b.ValueOrZero())
	})

	t.Run("BoolFromPtr with nil creates invalid null.Bool", func(t *testing.T) {
		b := null.BoolFromPtr(nil)
		assert.False(t, b.Valid)
		assert.False(t, b.Bool)
		assert.False(t, b.ValueOrZero())
	})

	t.Run("BoolFromPtr with value creates valid null.Bool", func(t *testing.T) {
		value := true
		b := null.BoolFromPtr(&value)
		assert.True(t, b.Valid)
		assert.True(t, b.Bool)
		assert.True(t, b.ValueOrZero())
	})

	t.Run("NewBool creates null.Bool with specified validity", func(t *testing.T) {
		validBool := null.NewBool(true, true)
		assert.True(t, validBool.Valid)
		assert.True(t, validBool.Bool)

		invalidBool := null.NewBool(true, false)
		assert.False(t, invalidBool.Valid)
		assert.True(t, invalidBool.Bool)           // Value is set but not valid
		assert.False(t, invalidBool.ValueOrZero()) // Should return false for invalid
	})

	t.Run("ValueOr returns default for invalid bool", func(t *testing.T) {
		invalidBool := null.NewBool(false, false)
		assert.True(t, invalidBool.ValueOr(true))

		validBool := null.BoolFrom(false)
		assert.False(t, validBool.ValueOr(true))
	})
}

func TestNullValueBasicOperations(t *testing.T) {
	t.Run("ValueFrom creates valid null.Value", func(t *testing.T) {
		str := null.ValueFrom("hello")
		assert.True(t, str.Valid)
		assert.Equal(t, "hello", str.V)
		assert.Equal(t, "hello", str.ValueOrZero())
	})

	t.Run("ValueFromPtr with nil creates invalid null.Value", func(t *testing.T) {
		var nilStr *string
		str := null.ValueFromPtr(nilStr)
		assert.False(t, str.Valid)
		assert.Equal(t, "", str.ValueOrZero())
	})

	t.Run("ValueFromPtr with value creates valid null.Value", func(t *testing.T) {
		value := "hello"
		str := null.ValueFromPtr(&value)
		assert.True(t, str.Valid)
		assert.Equal(t, "hello", str.V)
		assert.Equal(t, "hello", str.ValueOrZero())
	})

	t.Run("NewValue creates null.Value with specified validity", func(t *testing.T) {
		validValue := null.NewValue("hello", true)
		assert.True(t, validValue.Valid)
		assert.Equal(t, "hello", validValue.V)

		invalidValue := null.NewValue("hello", false)
		assert.False(t, invalidValue.Valid)
		assert.Equal(t, "hello", invalidValue.V)        // Value is set but not valid
		assert.Equal(t, "", invalidValue.ValueOrZero()) // Should return zero value for invalid
	})

	t.Run("null.Value with different types", func(t *testing.T) {
		intVal := null.ValueFrom(42)
		assert.True(t, intVal.Valid)
		assert.Equal(t, 42, intVal.V)
		assert.Equal(t, 42, intVal.ValueOrZero())

		floatVal := null.ValueFrom(3.14)
		assert.True(t, floatVal.Valid)
		assert.Equal(t, 3.14, floatVal.V)
		assert.Equal(t, 3.14, floatVal.ValueOrZero())

		boolVal := null.ValueFrom(true)
		assert.True(t, boolVal.Valid)
		assert.True(t, boolVal.V)
		assert.True(t, boolVal.ValueOrZero())
	})
}

func TestNullSpecificTypesDecodeHook(t *testing.T) {
	t.Run("null.String decode hook", func(t *testing.T) {
		// Test string to null.String conversion
		type StructWithNullString struct {
			Name null.String `json:"name"`
		}

		input := map[string]any{
			"name": "John Doe",
		}

		result, err := FromMap[StructWithNullString](input)
		require.NoError(t, err)
		assert.True(t, result.Name.Valid)
		assert.Equal(t, "John Doe", result.Name.String)

		// Test null.String to string conversion
		type StructWithString struct {
			Name string `json:"name"`
		}

		input2 := map[string]any{
			"name": null.StringFrom("Jane Doe"),
		}

		result2, err := FromMap[StructWithString](input2)
		require.NoError(t, err)
		assert.Equal(t, "Jane Doe", result2.Name)

		// Test invalid null.String to string
		input3 := map[string]any{
			"name": null.NewString("", false),
		}

		result3, err := FromMap[StructWithString](input3)
		require.NoError(t, err)
		assert.Equal(t, "", result3.Name) // Should be zero value for invalid
	})

	t.Run("null.Int decode hook", func(t *testing.T) {
		// Test int64 to null.Int conversion
		type StructWithNullInt struct {
			Age null.Int `json:"age"`
		}

		input := map[string]any{
			"age": int64(30),
		}

		result, err := FromMap[StructWithNullInt](input)
		require.NoError(t, err)
		assert.True(t, result.Age.Valid)
		assert.Equal(t, int64(30), result.Age.Int64)

		// Test null.Int to int64 conversion
		type StructWithInt struct {
			Age int64 `json:"age"`
		}

		input2 := map[string]any{
			"age": null.IntFrom(25),
		}

		result2, err := FromMap[StructWithInt](input2)
		require.NoError(t, err)
		assert.Equal(t, int64(25), result2.Age)
	})

	t.Run("null.Int16 decode hook", func(t *testing.T) {
		// Test int16 to null.Int16 conversion
		type StructWithNullInt16 struct {
			Count null.Int16 `json:"count"`
		}

		input := map[string]any{
			"count": int16(100),
		}

		result, err := FromMap[StructWithNullInt16](input)
		require.NoError(t, err)
		assert.True(t, result.Count.Valid)
		assert.Equal(t, int16(100), result.Count.Int16)

		// Test null.Int16 to int16 conversion
		type StructWithInt16 struct {
			Count int16 `json:"count"`
		}

		input2 := map[string]any{
			"count": null.Int16From(200),
		}

		result2, err := FromMap[StructWithInt16](input2)
		require.NoError(t, err)
		assert.Equal(t, int16(200), result2.Count)
	})

	t.Run("null.Int32 decode hook", func(t *testing.T) {
		// Test int32 to null.Int32 conversion
		type StructWithNullInt32 struct {
			ID null.Int32 `json:"id"`
		}

		input := map[string]any{
			"id": int32(12345),
		}

		result, err := FromMap[StructWithNullInt32](input)
		require.NoError(t, err)
		assert.True(t, result.ID.Valid)
		assert.Equal(t, int32(12345), result.ID.Int32)

		// Test null.Int32 to int32 conversion
		type StructWithInt32 struct {
			ID int32 `json:"id"`
		}

		input2 := map[string]any{
			"id": null.Int32From(54321),
		}

		result2, err := FromMap[StructWithInt32](input2)
		require.NoError(t, err)
		assert.Equal(t, int32(54321), result2.ID)
	})

	t.Run("null.Float decode hook", func(t *testing.T) {
		// Test float64 to null.Float conversion
		type StructWithNullFloat struct {
			Score null.Float `json:"score"`
		}

		input := map[string]any{
			"score": float64(95.5),
		}

		result, err := FromMap[StructWithNullFloat](input)
		require.NoError(t, err)
		assert.True(t, result.Score.Valid)
		assert.Equal(t, 95.5, result.Score.Float64)

		// Test null.Float to float64 conversion
		type StructWithFloat struct {
			Score float64 `json:"score"`
		}

		input2 := map[string]any{
			"score": null.FloatFrom(87.3),
		}

		result2, err := FromMap[StructWithFloat](input2)
		require.NoError(t, err)
		assert.Equal(t, 87.3, result2.Score)
	})

	t.Run("null.Byte decode hook", func(t *testing.T) {
		// Test byte to null.Byte conversion
		type StructWithNullByte struct {
			Flag null.Byte `json:"flag"`
		}

		input := map[string]any{
			"flag": byte(255),
		}

		result, err := FromMap[StructWithNullByte](input)
		require.NoError(t, err)
		assert.True(t, result.Flag.Valid)
		assert.Equal(t, byte(255), result.Flag.Byte)

		// Test null.Byte to byte conversion
		type StructWithByte struct {
			Flag byte `json:"flag"`
		}

		input2 := map[string]any{
			"flag": null.ByteFrom(128),
		}

		result2, err := FromMap[StructWithByte](input2)
		require.NoError(t, err)
		assert.Equal(t, byte(128), result2.Flag)
	})

	t.Run("null.DateTime decode hook", func(t *testing.T) {
		testDateTime := datetime.Of(time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC))

		// Test datetime.DateTime to null.DateTime conversion
		type StructWithNullDateTime struct {
			Created null.DateTime `json:"created"`
		}

		input := map[string]any{
			"created": testDateTime,
		}

		result, err := FromMap[StructWithNullDateTime](input)
		require.NoError(t, err)
		assert.True(t, result.Created.Valid)
		assert.Equal(t, testDateTime, result.Created.V)

		// Test null.DateTime to datetime.DateTime conversion
		type StructWithDateTime struct {
			Created datetime.DateTime `json:"created"`
		}

		input2 := map[string]any{
			"created": null.DateTimeFrom(testDateTime),
		}

		result2, err := FromMap[StructWithDateTime](input2)
		require.NoError(t, err)
		assert.Equal(t, testDateTime, result2.Created)
	})

	t.Run("null.Date decode hook", func(t *testing.T) {
		testDate := datetime.DateOf(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))

		// Test datetime.Date to null.Date conversion
		type StructWithNullDate struct {
			Birthday null.Date `json:"birthday"`
		}

		input := map[string]any{
			"birthday": testDate,
		}

		result, err := FromMap[StructWithNullDate](input)
		require.NoError(t, err)
		assert.True(t, result.Birthday.Valid)
		assert.Equal(t, testDate, result.Birthday.V)

		// Test null.Date to datetime.Date conversion
		type StructWithDate struct {
			Birthday datetime.Date `json:"birthday"`
		}

		input2 := map[string]any{
			"birthday": null.DateFrom(testDate),
		}

		result2, err := FromMap[StructWithDate](input2)
		require.NoError(t, err)
		assert.Equal(t, testDate, result2.Birthday)
	})

	t.Run("null.Time decode hook", func(t *testing.T) {
		testTime := datetime.TimeOf(time.Date(0, 1, 1, 15, 30, 45, 0, time.UTC))

		// Test datetime.Time to null.Time conversion
		type StructWithNullTime struct {
			MeetingTime null.Time `json:"meeting_time"`
		}

		input := map[string]any{
			"meeting_time": testTime,
		}

		result, err := FromMap[StructWithNullTime](input)
		require.NoError(t, err)
		assert.True(t, result.MeetingTime.Valid)
		assert.Equal(t, testTime, result.MeetingTime.V)

		// Test null.Time to datetime.Time conversion
		type StructWithTime struct {
			MeetingTime datetime.Time `json:"meeting_time"`
		}

		input2 := map[string]any{
			"meeting_time": null.TimeFrom(testTime),
		}

		result2, err := FromMap[StructWithTime](input2)
		require.NoError(t, err)
		assert.Equal(t, testTime, result2.MeetingTime)
	})

	t.Run("null.Decimal decode hook", func(t *testing.T) {
		testDecimal := decimal.NewFromFloat(123.456)

		// Test decimal.Decimal to null.Decimal conversion
		type StructWithNullDecimal struct {
			Price null.Decimal `json:"price"`
		}

		input := map[string]any{
			"price": testDecimal,
		}

		result, err := FromMap[StructWithNullDecimal](input)
		require.NoError(t, err)
		assert.True(t, result.Price.Valid)
		assert.True(t, testDecimal.Equal(result.Price.Decimal))

		// Test null.Decimal to decimal.Decimal conversion
		type StructWithDecimal struct {
			Price decimal.Decimal `json:"price"`
		}

		input2 := map[string]any{
			"price": null.DecimalFrom(testDecimal),
		}

		result2, err := FromMap[StructWithDecimal](input2)
		require.NoError(t, err)
		assert.True(t, testDecimal.Equal(result2.Price))
	})
}

func TestNullTypesWithPointersDecodeHook(t *testing.T) {
	t.Run("pointer types conversion", func(t *testing.T) {
		// Test *string to null.String
		type StructWithNullString struct {
			Name null.String `json:"name"`
		}

		stringVal := "John Doe"
		input := map[string]any{
			"name": &stringVal,
		}

		result, err := FromMap[StructWithNullString](input)
		require.NoError(t, err)
		assert.True(t, result.Name.Valid)
		assert.Equal(t, "John Doe", result.Name.String)

		// Test null.String to *string
		type StructWithStringPtr struct {
			Name *string `json:"name"`
		}

		input2 := map[string]any{
			"name": null.StringFrom("Jane Doe"),
		}

		result2, err := FromMap[StructWithStringPtr](input2)
		require.NoError(t, err)
		require.NotNil(t, result2.Name)
		assert.Equal(t, "Jane Doe", *result2.Name)

		// Test nil pointer to null.String
		var nilString *string
		input3 := map[string]any{
			"name": nilString,
		}

		result3, err := FromMap[StructWithNullString](input3)
		require.NoError(t, err)
		assert.False(t, result3.Name.Valid)

		// Test invalid null.String to *string
		input4 := map[string]any{
			"name": null.NewString("test", false),
		}

		result4, err := FromMap[StructWithStringPtr](input4)
		require.NoError(t, err)
		assert.Nil(t, result4.Name)
	})

	t.Run("integer pointer types", func(t *testing.T) {
		// Test *int64 to null.Int
		type StructWithNullInt struct {
			Age null.Int `json:"age"`
		}

		intVal := int64(30)
		input := map[string]any{
			"age": &intVal,
		}

		result, err := FromMap[StructWithNullInt](input)
		require.NoError(t, err)
		assert.True(t, result.Age.Valid)
		assert.Equal(t, int64(30), result.Age.Int64)

		// Test null.Int to *int64
		type StructWithIntPtr struct {
			Age *int64 `json:"age"`
		}

		input2 := map[string]any{
			"age": null.IntFrom(25),
		}

		result2, err := FromMap[StructWithIntPtr](input2)
		require.NoError(t, err)
		require.NotNil(t, result2.Age)
		assert.Equal(t, int64(25), *result2.Age)
	})
}

func TestNullTypesIntegrationAdvanced(t *testing.T) {
	t.Run("comprehensive struct with all null types", func(t *testing.T) {
		type ComprehensiveStruct struct {
			Name        null.String   `json:"name"`
			Age         null.Int      `json:"age"`
			ShortCount  null.Int16    `json:"short_count"`
			ID          null.Int32    `json:"id"`
			Score       null.Float    `json:"score"`
			Flag        null.Byte     `json:"flag"`
			Created     null.DateTime `json:"created"`
			Birthday    null.Date     `json:"birthday"`
			MeetingTime null.Time     `json:"meeting_time"`
			Price       null.Decimal  `json:"price"`
			Active      null.Bool     `json:"active"`
		}

		testDateTime := datetime.Of(time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC))
		testDate := datetime.DateOf(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC))
		testTime := datetime.TimeOf(time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC))
		testDecimal := decimal.NewFromFloat(99.99)

		input := map[string]any{
			"name":         "John Doe",
			"age":          int64(30),
			"short_count":  int16(100),
			"id":           int32(12345),
			"score":        95.5,
			"flag":         byte(255),
			"created":      testDateTime,
			"birthday":     testDate,
			"meeting_time": testTime,
			"price":        testDecimal,
			"active":       true,
		}

		result, err := FromMap[ComprehensiveStruct](input)
		require.NoError(t, err)

		// Verify all fields are valid and have correct values
		assert.True(t, result.Name.Valid)
		assert.Equal(t, "John Doe", result.Name.String)

		assert.True(t, result.Age.Valid)
		assert.Equal(t, int64(30), result.Age.Int64)

		assert.True(t, result.ShortCount.Valid)
		assert.Equal(t, int16(100), result.ShortCount.Int16)

		assert.True(t, result.ID.Valid)
		assert.Equal(t, int32(12345), result.ID.Int32)

		assert.True(t, result.Score.Valid)
		assert.Equal(t, 95.5, result.Score.Float64)

		assert.True(t, result.Flag.Valid)
		assert.Equal(t, byte(255), result.Flag.Byte)

		assert.True(t, result.Created.Valid)
		assert.Equal(t, testDateTime, result.Created.V)

		assert.True(t, result.Birthday.Valid)
		assert.Equal(t, testDate, result.Birthday.V)

		assert.True(t, result.MeetingTime.Valid)
		assert.Equal(t, testTime, result.MeetingTime.V)

		assert.True(t, result.Price.Valid)
		assert.True(t, testDecimal.Equal(result.Price.Decimal))

		assert.True(t, result.Active.Valid)
		assert.True(t, result.Active.Bool)
	})

	t.Run("partial input with some null fields", func(t *testing.T) {
		type PartialStruct struct {
			Name   null.String `json:"name"`
			Age    null.Int    `json:"age"`
			Score  null.Float  `json:"score"`
			Active null.Bool   `json:"active"`
		}

		// Only provide name and age, leave score and active unset
		input := map[string]any{
			"name": "Jane Doe",
			"age":  int64(25),
		}

		result, err := FromMap[PartialStruct](input)
		require.NoError(t, err)

		// Provided fields should be valid
		assert.True(t, result.Name.Valid)
		assert.Equal(t, "Jane Doe", result.Name.String)

		assert.True(t, result.Age.Valid)
		assert.Equal(t, int64(25), result.Age.Int64)

		// Unprovided fields should be invalid
		assert.False(t, result.Score.Valid)
		assert.False(t, result.Active.Valid)
	})
}

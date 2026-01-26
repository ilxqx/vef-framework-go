package encoding

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type JSONTestSuite struct {
	suite.Suite
}

func TestJSONTestSuite(t *testing.T) {
	suite.Run(t, new(JSONTestSuite))
}

func (suite *JSONTestSuite) TestToJSON() {
	suite.Run("ValidStruct", func() {
		input := generateSimpleStruct()
		result, err := ToJSON(input)

		require.NoError(suite.T(), err, "ToJSON should encode valid struct without error")
		assert.Contains(suite.T(), result, `"name"`, "JSON should contain name field")
		assert.Contains(suite.T(), result, `"age"`, "JSON should contain age field")
		assert.Contains(suite.T(), result, `"active"`, "JSON should contain active field")
	})

	suite.Run("NilInput", func() {
		result, err := ToJSON(nil)

		require.NoError(suite.T(), err, "ToJSON should handle nil input without error")
		assert.Equal(suite.T(), "null", result, "ToJSON should encode nil as 'null'")
	})

	suite.Run("EmptyStruct", func() {
		input := SimpleStruct{}
		result, err := ToJSON(input)

		require.NoError(suite.T(), err, "ToJSON should encode empty struct without error")
		assert.Contains(suite.T(), result, `"name":""`, "JSON should contain empty name field")
		assert.Contains(suite.T(), result, `"age":0`, "JSON should contain zero age field")
		assert.Contains(suite.T(), result, `"active":false`, "JSON should contain false active field")
	})

	suite.Run("StructWithAllFields", func() {
		input := generateMediumStruct(3)
		result, err := ToJSON(input)

		require.NoError(suite.T(), err, "ToJSON should encode complex struct without error")
		assert.Contains(suite.T(), result, `"id"`, "JSON should contain id field")
		assert.Contains(suite.T(), result, `"items"`, "JSON should contain items field")
		assert.Contains(suite.T(), result, `"tags"`, "JSON should contain tags field")
	})

	suite.Run("StructWithOmitEmpty", func() {
		input := SimpleStruct{
			Name:   "Charlie",
			Age:    25,
			Active: true,
		}
		result, err := ToJSON(input)

		require.NoError(suite.T(), err, "ToJSON should encode struct with omitempty without error")
		assertNotEmptyWithContext(suite.T(), result, "JSON output should not be empty")
	})
}

func (suite *JSONTestSuite) TestFromJSON() {
	suite.Run("ValidJSON", func() {
		input := `{"id":"test123","items":[],"tags":["tag1","tag2"],"metadata":{"key":"value"}}`
		result, err := FromJSON[MediumStruct](input)

		require.NoError(suite.T(), err, "FromJSON should decode valid JSON without error")
		require.NotNil(suite.T(), result, "FromJSON should return non-nil result")
		assert.Equal(suite.T(), "test123", result.ID, "ID field should match input")
		assert.NotNil(suite.T(), result.Items, "Items field should not be nil")
		assert.Len(suite.T(), result.Tags, 2, "Tags should have 2 elements")
		assert.NotNil(suite.T(), result.Metadata, "Metadata field should not be nil")
	})

	suite.Run("PartialJSON", func() {
		input := `{"id":"partial"}`
		result, err := FromJSON[MediumStruct](input)

		require.NoError(suite.T(), err, "FromJSON should decode partial JSON without error")
		require.NotNil(suite.T(), result, "FromJSON should return non-nil result")
		assert.Equal(suite.T(), "partial", result.ID, "ID field should match input")
		assert.Nil(suite.T(), result.Items, "Items field should be nil")
		assert.Nil(suite.T(), result.Tags, "Tags field should be nil")
	})

	suite.Run("InvalidJSON", func() {
		input := `{"id":"test","items":}`
		_, err := FromJSON[MediumStruct](input)

		assertErrorWithContext(suite.T(), err, "FromJSON should return error for malformed JSON")
	})

	suite.Run("EmptyJSON", func() {
		input := `{}`
		result, err := FromJSON[MediumStruct](input)

		require.NoError(suite.T(), err, "FromJSON should decode empty JSON object without error")
		require.NotNil(suite.T(), result, "FromJSON should return non-nil result")
		assert.Equal(suite.T(), "", result.ID, "ID field should be empty string")
		assert.Nil(suite.T(), result.Items, "Items field should be nil")
		assert.Nil(suite.T(), result.Tags, "Tags field should be nil")
	})

	suite.Run("JSONWithExtraFields", func() {
		input := `{"id":"test","items":[],"extra":"field","unknown":123}`
		result, err := FromJSON[MediumStruct](input)

		require.NoError(suite.T(), err, "FromJSON should ignore extra fields without error")
		require.NotNil(suite.T(), result, "FromJSON should return non-nil result")
		assert.Equal(suite.T(), "test", result.ID, "ID field should match input")
	})

	suite.Run("JSONWithUnicodeCharacters", func() {
		input := `{"id":"测试ID","tags":["标签1","标签2"]}`
		result, err := FromJSON[MediumStruct](input)

		require.NoError(suite.T(), err, "FromJSON should decode Unicode characters without error")
		require.NotNil(suite.T(), result, "FromJSON should return non-nil result")
		assert.Equal(suite.T(), "测试ID", result.ID, "ID field should preserve Unicode characters")
		assert.Len(suite.T(), result.Tags, 2, "Tags should have 2 elements")
	})
}

func (suite *JSONTestSuite) TestDecodeJSON() {
	suite.Run("DecodeIntoStructPointer", func() {
		input := `{"id":"test123","items":[],"tags":["tag1"]}`

		var result MediumStruct

		err := DecodeJSON(input, &result)

		require.NoError(suite.T(), err, "DecodeJSON should decode into struct pointer without error")
		assert.Equal(suite.T(), "test123", result.ID, "ID field should match input")
		assert.NotNil(suite.T(), result.Items, "Items field should not be nil")
		assert.Len(suite.T(), result.Tags, 1, "Tags should have 1 element")
	})

	suite.Run("InvalidJSON", func() {
		input := `{"id":"test","items":}`

		var result MediumStruct

		err := DecodeJSON(input, &result)

		assertErrorWithContext(suite.T(), err, "DecodeJSON should return error for malformed JSON")
	})

	suite.Run("EmptyJSON", func() {
		input := `{}`

		var result MediumStruct

		err := DecodeJSON(input, &result)

		require.NoError(suite.T(), err, "DecodeJSON should decode empty JSON object without error")
		assert.Equal(suite.T(), "", result.ID, "ID field should be empty string")
	})

	suite.Run("DecodeIntoNonPointer", func() {
		input := `{"id":"test"}`

		var result MediumStruct

		err := DecodeJSON(input, result)

		assertErrorWithContext(suite.T(), err, "DecodeJSON should return error when target is not a pointer")
	})
}

func (suite *JSONTestSuite) TestJSONRoundTrip() {
	suite.Run("SimpleStruct", func() {
		input := generateSimpleStruct()

		encoded, err := ToJSON(input)
		require.NoError(suite.T(), err, "ToJSON should encode struct without error")
		assertNotEmptyWithContext(suite.T(), encoded, "Encoded JSON should not be empty")

		decoded, err := FromJSON[SimpleStruct](encoded)
		require.NoError(suite.T(), err, "FromJSON should decode JSON without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assertStructEqual(suite.T(), input, *decoded, "Round-trip should preserve all fields")
	})

	suite.Run("MediumStruct", func() {
		input := generateMediumStruct(3)

		encoded, err := ToJSON(input)
		require.NoError(suite.T(), err, "ToJSON should encode struct without error")
		assertNotEmptyWithContext(suite.T(), encoded, "Encoded JSON should not be empty")

		decoded, err := FromJSON[MediumStruct](encoded)
		require.NoError(suite.T(), err, "FromJSON should decode JSON without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assert.Equal(suite.T(), input.ID, decoded.ID, "ID field should be preserved")
		assert.Equal(suite.T(), len(input.Items), len(decoded.Items), "Items length should be preserved")
		assert.Equal(suite.T(), input.Tags, decoded.Tags, "Tags slice should be preserved")
		assert.Equal(suite.T(), input.Metadata, decoded.Metadata, "Metadata map should be preserved")
	})

	suite.Run("ComplexStruct", func() {
		input := generateComplexStruct(1)

		encoded, err := ToJSON(input)
		require.NoError(suite.T(), err, "ToJSON should encode complex struct without error")
		assertNotEmptyWithContext(suite.T(), encoded, "Encoded JSON should not be empty")

		decoded, err := FromJSON[ComplexStruct](encoded)
		require.NoError(suite.T(), err, "FromJSON should decode JSON without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assert.Equal(suite.T(), input.ID, decoded.ID, "ID field should be preserved")
		assert.Equal(suite.T(), input.Score, decoded.Score, "Score field should be preserved")
		assert.Equal(suite.T(), input.Count, decoded.Count, "Count field should be preserved")
		assert.Equal(suite.T(), input.Enabled, decoded.Enabled, "Enabled field should be preserved")
		assert.Equal(suite.T(), input.Created.Unix(), decoded.Created.Unix(), "Created timestamp should be preserved")
	})

	suite.Run("StructWithTimeField", func() {
		created := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		input := ComplexStruct{
			ID:      "time-test",
			Data:    map[string]any{"test": true},
			Items:   []MediumStruct{},
			Created: created,
			Score:   99.5,
			Count:   100,
			Enabled: true,
		}

		encoded, err := ToJSON(input)
		require.NoError(suite.T(), err, "ToJSON should encode struct with time field without error")

		decoded, err := FromJSON[ComplexStruct](encoded)
		require.NoError(suite.T(), err, "FromJSON should decode struct with time field without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assert.Equal(suite.T(), input.Created.Unix(), decoded.Created.Unix(), "Time field should be preserved (comparing Unix timestamps)")
	})
}

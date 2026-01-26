package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GOBTestSuite struct {
	suite.Suite
}

func TestGOBTestSuite(t *testing.T) {
	suite.Run(t, new(GOBTestSuite))
}

func (suite *GOBTestSuite) TestToGOB() {
	suite.Run("ValidStruct", func() {
		input := generateSimpleStruct()
		result, err := ToGOB(input)

		require.NoError(suite.T(), err, "ToGOB should encode valid struct without error")
		require.NotNil(suite.T(), result, "ToGOB should return non-nil byte slice")
		assertNotEmptyWithContext(suite.T(), result, "Encoded GOB data should not be empty")
	})

	suite.Run("EmptyStruct", func() {
		input := SimpleStruct{}
		result, err := ToGOB(input)

		require.NoError(suite.T(), err, "ToGOB should encode empty struct without error")
		require.NotNil(suite.T(), result, "ToGOB should return non-nil byte slice")
		assertNotEmptyWithContext(suite.T(), result, "Encoded GOB data should not be empty even for empty struct")
	})

	suite.Run("ComplexStruct", func() {
		input := generateComplexStruct(1)
		result, err := ToGOB(input)

		require.NoError(suite.T(), err, "ToGOB should encode complex struct without error")
		require.NotNil(suite.T(), result, "ToGOB should return non-nil byte slice")
		assertNotEmptyWithContext(suite.T(), result, "Encoded GOB data should not be empty")
	})

	suite.Run("StructWithSlices", func() {
		input := generateMediumStruct(3)
		result, err := ToGOB(input)

		require.NoError(suite.T(), err, "ToGOB should encode struct with slices without error")
		require.NotNil(suite.T(), result, "ToGOB should return non-nil byte slice")
		assertNotEmptyWithContext(suite.T(), result, "Encoded GOB data should not be empty")
	})
}

func (suite *GOBTestSuite) TestFromGOB() {
	suite.Run("ValidData", func() {
		input := generateMediumStruct(3)
		data, err := ToGOB(input)
		require.NoError(suite.T(), err, "ToGOB should succeed for test setup")

		result, err := FromGOB[MediumStruct](data)

		require.NoError(suite.T(), err, "FromGOB should decode valid GOB data without error")
		require.NotNil(suite.T(), result, "FromGOB should return non-nil result")
		assert.Equal(suite.T(), input.ID, result.ID, "ID field should match input")
		assert.Equal(suite.T(), len(input.Items), len(result.Items), "Items length should match input")
		assert.Equal(suite.T(), input.Tags, result.Tags, "Tags should match input")
	})

	suite.Run("InvalidData", func() {
		data := []byte("invalid gob data")
		_, err := FromGOB[MediumStruct](data)

		assertErrorWithContext(suite.T(), err, "FromGOB should return error for invalid GOB data")
	})

	suite.Run("EmptyData", func() {
		data := []byte{}
		_, err := FromGOB[MediumStruct](data)

		assertErrorWithContext(suite.T(), err, "FromGOB should return error for empty byte slice")
	})

	suite.Run("CorruptedData", func() {
		input := generateSimpleStruct()
		data, err := ToGOB(input)
		require.NoError(suite.T(), err, "ToGOB should succeed for test setup")

		// Truncate data to ensure decoding fails
		if len(data) > 5 {
			data = data[:5]
		}
		_, err = FromGOB[SimpleStruct](data)

		assertErrorWithContext(suite.T(), err, "FromGOB should return error for corrupted GOB data")
	})
}

func (suite *GOBTestSuite) TestDecodeGOB() {
	suite.Run("ValidData", func() {
		input := generateMediumStruct(3)
		data, err := ToGOB(input)
		require.NoError(suite.T(), err, "ToGOB should succeed for test setup")

		var result MediumStruct
		err = DecodeGOB(data, &result)

		require.NoError(suite.T(), err, "DecodeGOB should decode into struct pointer without error")
		assert.Equal(suite.T(), input.ID, result.ID, "ID field should match input")
		assert.Equal(suite.T(), len(input.Items), len(result.Items), "Items length should match input")
	})

	suite.Run("InvalidData", func() {
		data := []byte("invalid gob data")
		var result MediumStruct

		err := DecodeGOB(data, &result)

		assertErrorWithContext(suite.T(), err, "DecodeGOB should return error for invalid GOB data")
	})

	suite.Run("EmptyData", func() {
		data := []byte{}
		var result MediumStruct

		err := DecodeGOB(data, &result)

		assertErrorWithContext(suite.T(), err, "DecodeGOB should return error for empty byte slice")
	})

	suite.Run("DecodeIntoNonPointer", func() {
		input := generateSimpleStruct()
		data, err := ToGOB(input)
		require.NoError(suite.T(), err, "ToGOB should succeed for test setup")

		var result SimpleStruct
		err = DecodeGOB(data, result)

		assertErrorWithContext(suite.T(), err, "DecodeGOB should return error when target is not a pointer")
	})
}

func (suite *GOBTestSuite) TestGOBRoundTrip() {
	suite.Run("SimpleStruct", func() {
		input := generateSimpleStruct()

		encoded, err := ToGOB(input)
		require.NoError(suite.T(), err, "ToGOB should encode struct without error")
		assertNotEmptyWithContext(suite.T(), encoded, "Encoded GOB data should not be empty")

		decoded, err := FromGOB[SimpleStruct](encoded)
		require.NoError(suite.T(), err, "FromGOB should decode GOB data without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assertStructEqual(suite.T(), input, *decoded, "Round-trip should preserve all fields")
	})

	suite.Run("MediumStruct", func() {
		input := generateMediumStruct(3)

		encoded, err := ToGOB(input)
		require.NoError(suite.T(), err, "ToGOB should encode struct without error")
		assertNotEmptyWithContext(suite.T(), encoded, "Encoded GOB data should not be empty")

		decoded, err := FromGOB[MediumStruct](encoded)
		require.NoError(suite.T(), err, "FromGOB should decode GOB data without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assert.Equal(suite.T(), input.ID, decoded.ID, "ID field should be preserved")
		assert.Equal(suite.T(), len(input.Items), len(decoded.Items), "Items length should be preserved")
		assert.Equal(suite.T(), input.Tags, decoded.Tags, "Tags slice should be preserved")
		assert.Equal(suite.T(), input.Metadata, decoded.Metadata, "Metadata map should be preserved")
	})

	suite.Run("ComplexStruct", func() {
		input := generateComplexStruct(1)

		encoded, err := ToGOB(input)
		require.NoError(suite.T(), err, "ToGOB should encode complex struct without error")
		assertNotEmptyWithContext(suite.T(), encoded, "Encoded GOB data should not be empty")

		decoded, err := FromGOB[ComplexStruct](encoded)
		require.NoError(suite.T(), err, "FromGOB should decode GOB data without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assert.Equal(suite.T(), input.ID, decoded.ID, "ID field should be preserved")
		assert.Equal(suite.T(), input.Score, decoded.Score, "Score field should be preserved")
		assert.Equal(suite.T(), input.Count, decoded.Count, "Count field should be preserved")
		assert.Equal(suite.T(), input.Enabled, decoded.Enabled, "Enabled field should be preserved")
		assert.Equal(suite.T(), input.Created.Unix(), decoded.Created.Unix(), "Created timestamp should be preserved")
	})

	suite.Run("LargeData", func() {
		input := generateLargeStruct(10240)

		encoded, err := ToGOB(input)
		require.NoError(suite.T(), err, "ToGOB should encode large struct without error")
		assertNotEmptyWithContext(suite.T(), encoded, "Encoded GOB data should not be empty")

		decoded, err := FromGOB[ComplexStruct](encoded)
		require.NoError(suite.T(), err, "FromGOB should decode large GOB data without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assert.Equal(suite.T(), input.ID, decoded.ID, "ID field should be preserved")
		assert.Equal(suite.T(), len(input.Items), len(decoded.Items), "Items slice length should be preserved")
	})
}

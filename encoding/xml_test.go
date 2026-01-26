package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type XMLTestSuite struct {
	suite.Suite
}

func TestXMLTestSuite(t *testing.T) {
	suite.Run(t, new(XMLTestSuite))
}

func (suite *XMLTestSuite) TestToXML() {
	suite.Run("ValidStruct", func() {
		input := generateSimpleStruct()
		result, err := ToXML(input)

		require.NoError(suite.T(), err, "ToXML should encode valid struct without error")
		assert.Contains(suite.T(), result, "<SimpleStruct>", "XML should contain root element")
		assert.Contains(suite.T(), result, "<Name>", "XML should contain Name element")
		assert.Contains(suite.T(), result, "<Age>", "XML should contain Age element")
		assert.Contains(suite.T(), result, "<Active>", "XML should contain Active element")
	})

	suite.Run("EmptyStruct", func() {
		input := SimpleStruct{}
		result, err := ToXML(input)

		require.NoError(suite.T(), err, "ToXML should encode empty struct without error")
		assert.Contains(suite.T(), result, "<SimpleStruct>", "XML should contain root element")
		assert.Contains(suite.T(), result, "<Name></Name>", "XML should contain empty Name element")
		assert.Contains(suite.T(), result, "<Age>0</Age>", "XML should contain zero Age element")
		assert.Contains(suite.T(), result, "<Active>false</Active>", "XML should contain false Active element")
	})

	suite.Run("StructWithSpecialCharacters", func() {
		input := SimpleStruct{
			Name:   "Test & <Special>",
			Age:    30,
			Active: true,
		}
		result, err := ToXML(input)

		require.NoError(suite.T(), err, "ToXML should encode struct with special characters without error")
		assert.Contains(suite.T(), result, "&amp;", "XML should escape ampersand")
		assert.Contains(suite.T(), result, "&lt;", "XML should escape less-than sign")
		assert.Contains(suite.T(), result, "&gt;", "XML should escape greater-than sign")
	})
}

func (suite *XMLTestSuite) TestFromXML() {
	suite.Run("ValidXML", func() {
		input := `<SimpleStruct><Name>John Doe</Name><Age>30</Age><Active>true</Active></SimpleStruct>`
		result, err := FromXML[SimpleStruct](input)

		require.NoError(suite.T(), err, "FromXML should decode valid XML without error")
		require.NotNil(suite.T(), result, "FromXML should return non-nil result")
		assert.Equal(suite.T(), "John Doe", result.Name, "Name field should match input")
		assert.Equal(suite.T(), 30, result.Age, "Age field should match input")
		assert.True(suite.T(), result.Active, "Active field should be true")
	})

	suite.Run("PartialXML", func() {
		input := `<SimpleStruct><Name>Jane Doe</Name></SimpleStruct>`
		result, err := FromXML[SimpleStruct](input)

		require.NoError(suite.T(), err, "FromXML should decode partial XML without error")
		require.NotNil(suite.T(), result, "FromXML should return non-nil result")
		assert.Equal(suite.T(), "Jane Doe", result.Name, "Name field should match input")
		assert.Equal(suite.T(), 0, result.Age, "Age field should have zero value")
		assert.False(suite.T(), result.Active, "Active field should have false value")
	})

	suite.Run("InvalidXMLMissingClosingTag", func() {
		input := `<SimpleStruct><Name>John Doe</Name><Age>30</SimpleStruct>`
		_, err := FromXML[SimpleStruct](input)

		assertErrorWithContext(suite.T(), err, "FromXML should return error for malformed XML")
	})

	suite.Run("EmptyXML", func() {
		input := `<SimpleStruct></SimpleStruct>`
		result, err := FromXML[SimpleStruct](input)

		require.NoError(suite.T(), err, "FromXML should decode empty XML element without error")
		require.NotNil(suite.T(), result, "FromXML should return non-nil result")
		assert.Equal(suite.T(), "", result.Name, "Name field should be empty string")
		assert.Equal(suite.T(), 0, result.Age, "Age field should be zero")
		assert.False(suite.T(), result.Active, "Active field should be false")
	})

	suite.Run("XMLWithEscapedCharacters", func() {
		input := `<SimpleStruct><Name>Test &amp; &lt;Special&gt;</Name><Age>30</Age><Active>true</Active></SimpleStruct>`
		result, err := FromXML[SimpleStruct](input)

		require.NoError(suite.T(), err, "FromXML should decode escaped characters without error")
		require.NotNil(suite.T(), result, "FromXML should return non-nil result")
		assert.Equal(suite.T(), "Test & <Special>", result.Name, "Name field should contain unescaped characters")
	})
}

func (suite *XMLTestSuite) TestDecodeXML() {
	suite.Run("DecodeIntoStructPointer", func() {
		input := `<SimpleStruct><Name>John Doe</Name><Age>30</Age><Active>true</Active></SimpleStruct>`
		var result SimpleStruct

		err := DecodeXML(input, &result)

		require.NoError(suite.T(), err, "DecodeXML should decode into struct pointer without error")
		assert.Equal(suite.T(), "John Doe", result.Name, "Name field should match input")
		assert.Equal(suite.T(), 30, result.Age, "Age field should match input")
		assert.True(suite.T(), result.Active, "Active field should be true")
	})

	suite.Run("InvalidXML", func() {
		input := `<SimpleStruct><Name>John Doe</Name><Age>30</SimpleStruct>`
		var result SimpleStruct

		err := DecodeXML(input, &result)

		assertErrorWithContext(suite.T(), err, "DecodeXML should return error for malformed XML")
	})

	suite.Run("DecodeIntoNonPointer", func() {
		input := `<SimpleStruct><Name>John Doe</Name><Age>30</Age></SimpleStruct>`
		var result SimpleStruct

		err := DecodeXML(input, result)

		assertErrorWithContext(suite.T(), err, "DecodeXML should return error when target is not a pointer")
	})
}

func (suite *XMLTestSuite) TestXMLRoundTrip() {
	suite.Run("SimpleStruct", func() {
		input := generateSimpleStruct()

		encoded, err := ToXML(input)
		require.NoError(suite.T(), err, "ToXML should encode struct without error")
		assertNotEmptyWithContext(suite.T(), encoded, "Encoded XML should not be empty")

		decoded, err := FromXML[SimpleStruct](encoded)
		require.NoError(suite.T(), err, "FromXML should decode XML without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assertStructEqual(suite.T(), input, *decoded, "Round-trip should preserve all fields")
	})

	suite.Run("StructWithSpecialCharacters", func() {
		input := SimpleStruct{
			Name:   "Test & <Special> \"Quotes\"",
			Age:    30,
			Active: true,
		}

		encoded, err := ToXML(input)
		require.NoError(suite.T(), err, "ToXML should encode struct with special characters without error")

		decoded, err := FromXML[SimpleStruct](encoded)
		require.NoError(suite.T(), err, "FromXML should decode XML with escaped characters without error")
		require.NotNil(suite.T(), decoded, "Decoded result should not be nil")

		assert.Equal(suite.T(), input.Name, decoded.Name, "Special characters should be preserved in round-trip")
	})
}


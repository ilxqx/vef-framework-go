package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToXml(t *testing.T) {
	tests := []struct {
		name     string
		input    TestStruct
		expected []string
	}{
		{
			name: "ValidStruct",
			input: TestStruct{
				Name:   "John Doe",
				Age:    30,
				Active: true,
			},
			expected: []string{"<TestStruct>", "<Name>John Doe</Name>", "<Age>30</Age>", "<Active>true</Active>"},
		},
		{
			name:     "EmptyStruct",
			input:    TestStruct{},
			expected: []string{"<TestStruct>", "<Name></Name>", "<Age>0</Age>", "<Active>false</Active>"},
		},
		{
			name: "StructWithAllFields",
			input: TestStruct{
				Name:   "Jane Doe",
				Age:    25,
				Email:  "jane@example.com",
				Active: true,
				Score:  95.5,
			},
			expected: []string{
				"<Name>Jane Doe</Name>",
				"<Age>25</Age>",
				"<Email>jane@example.com</Email>",
				"<Score>95.5</Score>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToXml(tt.input)
			require.NoError(t, err)

			for _, exp := range tt.expected {
				assert.Contains(t, result, exp)
			}
		})
	}
}

func TestFromXml(t *testing.T) {
	t.Run("ValidXML", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</Age><Active>true</Active><Score>95.5</Score></TestStruct>`
		result, err := FromXml[TestStruct](input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.True(t, result.Active)
		assert.Equal(t, 95.5, result.Score)
	})

	t.Run("PartialXML", func(t *testing.T) {
		input := `<TestStruct><Name>Jane Doe</Name></TestStruct>`
		result, err := FromXml[TestStruct](input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Jane Doe", result.Name)
		assert.Equal(t, 0, result.Age)
		assert.False(t, result.Active)
	})

	t.Run("InvalidXMLMissingClosingTag", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</TestStruct>`
		_, err := FromXml[TestStruct](input)
		assert.Error(t, err)
	})

	t.Run("EmptyXML", func(t *testing.T) {
		input := `<TestStruct></TestStruct>`
		result, err := FromXml[TestStruct](input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
		assert.False(t, result.Active)
	})

	t.Run("XMLWithExtraElements", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</Age><Extra>field</Extra></TestStruct>`
		result, err := FromXml[TestStruct](input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
	})
}

func TestDecodeXml(t *testing.T) {
	t.Run("DecodeIntoStructPointer", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</Age><Active>true</Active></TestStruct>`

		var result TestStruct

		err := DecodeXml(input, &result)
		require.NoError(t, err)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.True(t, result.Active)
	})

	t.Run("InvalidXML", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</TestStruct>`

		var result TestStruct

		err := DecodeXml(input, &result)
		assert.Error(t, err)
	})

	t.Run("EmptyXML", func(t *testing.T) {
		input := `<TestStruct></TestStruct>`

		var result TestStruct

		err := DecodeXml(input, &result)
		require.NoError(t, err)
		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

func TestXmlRoundTrip(t *testing.T) {
	input := TestStruct{
		Name:   "Jane Doe",
		Age:    25,
		Email:  "jane@example.com",
		Active: true,
		Score:  88.5,
	}

	encoded, err := ToXml(input)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded, err := FromXml[TestStruct](encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)
	assert.Equal(t, input.Name, decoded.Name)
	assert.Equal(t, input.Age, decoded.Age)
	assert.Equal(t, input.Email, decoded.Email)
	assert.Equal(t, input.Active, decoded.Active)
	assert.Equal(t, input.Score, decoded.Score)
}

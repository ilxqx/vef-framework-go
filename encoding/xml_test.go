package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToXML(t *testing.T) {
	t.Run("Valid struct", func(t *testing.T) {
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

	t.Run("Empty struct", func(t *testing.T) {
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
	t.Run("Valid XML", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</Age><Active>true</Active><Score>95.5</Score></TestStruct>`

		result, err := FromXML[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Equal(t, true, result.Active)
		assert.Equal(t, 95.5, result.Score)
	})

	t.Run("Partial XML", func(t *testing.T) {
		input := `<TestStruct><Name>Jane Doe</Name></TestStruct>`

		result, err := FromXML[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Jane Doe", result.Name)
		assert.Equal(t, 0, result.Age)
	})

	t.Run("Invalid XML", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</TestStruct>`

		result, err := FromXML[TestStruct](input)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Empty XML", func(t *testing.T) {
		input := `<TestStruct></TestStruct>`

		result, err := FromXML[TestStruct](input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

func TestDecodeXML(t *testing.T) {
	t.Run("Decode into struct pointer", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</Age><Active>true</Active></TestStruct>`

		var result TestStruct

		err := DecodeXML(input, &result)
		require.NoError(t, err)

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.True(t, result.Active)
	})

	t.Run("Invalid XML", func(t *testing.T) {
		input := `<TestStruct><Name>John Doe</Name><Age>30</TestStruct>`

		var result TestStruct

		err := DecodeXML(input, &result)
		assert.Error(t, err)
	})
}

package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToGOB(t *testing.T) {
	t.Run("Valid struct", func(t *testing.T) {
		input := TestStruct{
			Name:   "John Doe",
			Age:    30,
			Active: true,
		}

		result, err := ToGOB(input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotEmpty(t, result)
	})
}

func TestFromGOB(t *testing.T) {
	t.Run("Valid data", func(t *testing.T) {
		input := TestStruct{
			Name:   "John Doe",
			Age:    30,
			Active: true,
			Score:  95.5,
		}

		data, err := ToGOB(input)
		require.NoError(t, err)

		result, err := FromGOB[TestStruct](data)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, input.Name, result.Name)
		assert.Equal(t, input.Age, result.Age)
		assert.Equal(t, input.Active, result.Active)
		assert.Equal(t, input.Score, result.Score)
	})

	t.Run("Invalid data", func(t *testing.T) {
		data := []byte("invalid gob data")

		result, err := FromGOB[TestStruct](data)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestDecodeGOB(t *testing.T) {
	t.Run("Valid data", func(t *testing.T) {
		input := TestStruct{
			Name:   "John Doe",
			Age:    30,
			Active: true,
		}

		data, err := ToGOB(input)
		require.NoError(t, err)

		var result TestStruct

		err = DecodeGOB(data, &result)
		require.NoError(t, err)

		assert.Equal(t, input.Name, result.Name)
		assert.Equal(t, input.Age, result.Age)
		assert.Equal(t, input.Active, result.Active)
	})

	t.Run("Invalid data", func(t *testing.T) {
		data := []byte("invalid gob data")

		var result TestStruct

		err := DecodeGOB(data, &result)
		assert.Error(t, err)
	})
}

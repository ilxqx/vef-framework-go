package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToGob(t *testing.T) {
	tests := []struct {
		name  string
		input TestStruct
	}{
		{
			name: "ValidStruct",
			input: TestStruct{
				Name:   "John Doe",
				Age:    30,
				Active: true,
			},
		},
		{
			name:  "EmptyStruct",
			input: TestStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToGob(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFromGob(t *testing.T) {
	t.Run("ValidData", func(t *testing.T) {
		input := TestStruct{
			Name:   "John Doe",
			Age:    30,
			Active: true,
			Score:  95.5,
		}

		data, err := ToGob(input)
		require.NoError(t, err)

		result, err := FromGob[TestStruct](data)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, input.Name, result.Name)
		assert.Equal(t, input.Age, result.Age)
		assert.Equal(t, input.Active, result.Active)
		assert.Equal(t, input.Score, result.Score)
	})

	t.Run("InvalidData", func(t *testing.T) {
		data := []byte("invalid gob data")
		_, err := FromGob[TestStruct](data)
		assert.Error(t, err)
	})

	t.Run("EmptyData", func(t *testing.T) {
		data := []byte{}
		_, err := FromGob[TestStruct](data)
		assert.Error(t, err)
	})
}

func TestDecodeGob(t *testing.T) {
	t.Run("ValidData", func(t *testing.T) {
		input := TestStruct{
			Name:   "John Doe",
			Age:    30,
			Active: true,
		}

		data, err := ToGob(input)
		require.NoError(t, err)

		var result TestStruct

		err = DecodeGob(data, &result)
		require.NoError(t, err)
		assert.Equal(t, input.Name, result.Name)
		assert.Equal(t, input.Age, result.Age)
		assert.Equal(t, input.Active, result.Active)
	})

	t.Run("InvalidData", func(t *testing.T) {
		data := []byte("invalid gob data")

		var result TestStruct

		err := DecodeGob(data, &result)
		assert.Error(t, err)
	})

	t.Run("EmptyData", func(t *testing.T) {
		data := []byte{}

		var result TestStruct

		err := DecodeGob(data, &result)
		assert.Error(t, err)
	})
}

func TestGobRoundTrip(t *testing.T) {
	input := TestStruct{
		Name:   "Jane Doe",
		Age:    25,
		Email:  "jane@example.com",
		Active: true,
		Score:  88.5,
	}

	encoded, err := ToGob(input)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded, err := FromGob[TestStruct](encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)
	assert.Equal(t, input.Name, decoded.Name)
	assert.Equal(t, input.Age, decoded.Age)
	assert.Equal(t, input.Email, decoded.Email)
	assert.Equal(t, input.Active, decoded.Active)
	assert.Equal(t, input.Score, decoded.Score)
}

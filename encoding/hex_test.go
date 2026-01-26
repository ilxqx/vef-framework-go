package encoding

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToHex(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "EmptyData",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "SimpleBytes",
			input:    []byte{0x00, 0x01, 0x02, 0x03, 0xff},
			expected: "00010203ff",
		},
		{
			name:     "TextData",
			input:    []byte("Hello"),
			expected: "48656c6c6f",
		},
		{
			name:     "UTF8Text",
			input:    []byte("中文"),
			expected: "e4b8ade69687",
		},
		{
			name: "AllByteValues",
			input: []byte{
				0x00, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70,
				0x80, 0x90, 0xa0, 0xb0, 0xc0, 0xd0, 0xe0, 0xf0,
			},
			expected: "00102030405060708090a0b0c0d0e0f0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToHex(tt.input)
			assert.Equal(t, tt.expected, result, "Hex encoding should produce lowercase hexadecimal string")
		})
	}
}

func TestFromHex(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  []byte
		expectErr bool
	}{
		{
			name:      "EmptyString",
			input:     "",
			expected:  nil,
			expectErr: false,
		},
		{
			name:      "SimpleHex",
			input:     "00010203ff",
			expected:  []byte{0x00, 0x01, 0x02, 0x03, 0xff},
			expectErr: false,
		},
		{
			name:      "TextData",
			input:     "48656c6c6f",
			expected:  []byte("Hello"),
			expectErr: false,
		},
		{
			name:      "UTF8Text",
			input:     "e4b8ade69687",
			expected:  []byte("中文"),
			expectErr: false,
		},
		{
			name:      "UppercaseHex",
			input:     "ABCDEF",
			expected:  []byte{0xab, 0xcd, 0xef},
			expectErr: false,
		},
		{
			name:      "MixedCaseHex",
			input:     "AbCdEf",
			expected:  []byte{0xab, 0xcd, 0xef},
			expectErr: false,
		},
		{
			name:      "InvalidHexOddLength",
			input:     "abc",
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "InvalidHexNonHexChar",
			input:     "abcg",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromHex(tt.input)
			if tt.expectErr {
				assert.Error(t, err, "FromHex should return error for invalid hex string")
			} else {
				require.NoError(t, err, "FromHex should decode valid hex string without error")
				assert.Equal(t, tt.expected, result, "Decoded bytes should match expected value")
			}
		})
	}
}

func TestHexRoundTrip(t *testing.T) {
	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err, "Random data generation should succeed")

	encoded := ToHex(data)
	assert.NotEmpty(t, encoded, "Encoded hex string should not be empty")
	assert.Equal(t, 512, len(encoded), "Hex string length should be twice the byte length")

	decoded, err := FromHex(encoded)
	require.NoError(t, err, "Decoding hex string should succeed")
	assert.Equal(t, data, decoded, "Round-trip encoding/decoding should preserve original data")
}

func TestHexCaseInsensitive(t *testing.T) {
	data := []byte{0xab, 0xcd, 0xef}

	t.Run("EncodingIsLowercase", func(t *testing.T) {
		encoded := ToHex(data)
		assert.Equal(t, "abcdef", encoded, "ToHex should produce lowercase hex string")
	})

	t.Run("DecodeUppercase", func(t *testing.T) {
		decoded, err := FromHex("ABCDEF")
		require.NoError(t, err, "FromHex should accept uppercase hex string")
		assert.Equal(t, data, decoded, "Decoded bytes should match original data")
	})

	t.Run("DecodeLowercase", func(t *testing.T) {
		decoded, err := FromHex("abcdef")
		require.NoError(t, err, "FromHex should accept lowercase hex string")
		assert.Equal(t, data, decoded, "Decoded bytes should match original data")
	})

	t.Run("DecodeMixedCase", func(t *testing.T) {
		decoded, err := FromHex("AbCdEf")
		require.NoError(t, err, "FromHex should accept mixed-case hex string")
		assert.Equal(t, data, decoded, "Decoded bytes should match original data")
	})
}

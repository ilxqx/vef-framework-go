package encoding

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToHex(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "Empty data",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "Simple bytes",
			input:    []byte{0x00, 0x01, 0x02, 0x03, 0xff},
			expected: "00010203ff",
		},
		{
			name:     "Text data",
			input:    []byte("Hello"),
			expected: "48656c6c6f",
		},
		{
			name:     "UTF-8 text",
			input:    []byte("中文"),
			expected: "e4b8ade69687",
		},
		{
			name:     "All bytes",
			input:    []byte{0x00, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80, 0x90, 0xa0, 0xb0, 0xc0, 0xd0, 0xe0, 0xf0},
			expected: "00102030405060708090a0b0c0d0e0f0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToHex(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFromHex(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: []byte{},
			wantErr:  false,
		},
		{
			name:     "Simple hex",
			input:    "00010203ff",
			expected: []byte{0x00, 0x01, 0x02, 0x03, 0xff},
			wantErr:  false,
		},
		{
			name:     "Text data",
			input:    "48656c6c6f",
			expected: []byte("Hello"),
			wantErr:  false,
		},
		{
			name:     "UTF-8 text",
			input:    "e4b8ade69687",
			expected: []byte("中文"),
			wantErr:  false,
		},
		{
			name:     "Uppercase hex",
			input:    "ABCDEF",
			expected: []byte{0xab, 0xcd, 0xef},
			wantErr:  false,
		},
		{
			name:     "Mixed case hex",
			input:    "AbCdEf",
			expected: []byte{0xab, 0xcd, 0xef},
			wantErr:  false,
		},
		{
			name:     "Invalid hex - odd length",
			input:    "abc",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid hex - non-hex character",
			input:    "abcg",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FromHex(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestHexRoundTrip(t *testing.T) {
	// Generate random data
	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err)

	// Encode to hex
	encoded := ToHex(data)
	assert.NotEmpty(t, encoded)
	assert.Equal(t, 512, len(encoded)) // 256 bytes = 512 hex characters

	// Decode back
	decoded, err := FromHex(encoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

func TestHexCaseInsensitive(t *testing.T) {
	data := []byte{0xab, 0xcd, 0xef}

	// Encode (always lowercase)
	encoded := ToHex(data)
	assert.Equal(t, "abcdef", encoded)

	// Decode uppercase
	decodedUpper, err := FromHex("ABCDEF")
	require.NoError(t, err)
	assert.Equal(t, data, decodedUpper)

	// Decode lowercase
	decodedLower, err := FromHex("abcdef")
	require.NoError(t, err)
	assert.Equal(t, data, decodedLower)

	// Decode mixed case
	decodedMixed, err := FromHex("AbCdEf")
	require.NoError(t, err)
	assert.Equal(t, data, decodedMixed)
}

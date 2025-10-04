package encoding

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToBase64(t *testing.T) {
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
			name:     "Simple text",
			input:    []byte("Hello, World!"),
			expected: "SGVsbG8sIFdvcmxkIQ==",
		},
		{
			name:     "Binary data",
			input:    []byte{0x00, 0x01, 0x02, 0x03, 0xff},
			expected: "AAECA/8=",
		},
		{
			name:     "UTF-8 text",
			input:    []byte("中文测试"),
			expected: "5Lit5paH5rWL6K+V",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToBase64(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFromBase64(t *testing.T) {
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
			name:     "Simple text",
			input:    "SGVsbG8sIFdvcmxkIQ==",
			expected: []byte("Hello, World!"),
			wantErr:  false,
		},
		{
			name:     "Binary data",
			input:    "AAECA/8=",
			expected: []byte{0x00, 0x01, 0x02, 0x03, 0xff},
			wantErr:  false,
		},
		{
			name:     "UTF-8 text",
			input:    "5Lit5paH5rWL6K+V",
			expected: []byte("中文测试"),
			wantErr:  false,
		},
		{
			name:     "Invalid base64",
			input:    "invalid!!!",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FromBase64(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBase64RoundTrip(t *testing.T) {
	// Generate random data
	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err)

	// Encode to base64
	encoded := ToBase64(data)
	assert.NotEmpty(t, encoded)

	// Decode back
	decoded, err := FromBase64(encoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

func TestToBase64URL(t *testing.T) {
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
			name:     "Data with + and /",
			input:    []byte{0xfb, 0xff, 0xbf},
			expected: "-_-_",
		},
		{
			name:     "Simple text",
			input:    []byte("Hello, World!"),
			expected: "SGVsbG8sIFdvcmxkIQ==",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToBase64URL(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFromBase64URL(t *testing.T) {
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
			name:     "URL-safe characters",
			input:    "-_-_",
			expected: []byte{0xfb, 0xff, 0xbf},
			wantErr:  false,
		},
		{
			name:     "Simple text",
			input:    "SGVsbG8sIFdvcmxkIQ==",
			expected: []byte("Hello, World!"),
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FromBase64URL(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBase64URLRoundTrip(t *testing.T) {
	// Generate random data
	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err)

	// Encode to base64 URL
	encoded := ToBase64URL(data)
	assert.NotEmpty(t, encoded)

	// Decode back
	decoded, err := FromBase64URL(encoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

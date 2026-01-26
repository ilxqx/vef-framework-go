package encoding

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToBase64(t *testing.T) {
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
			name:     "SimpleText",
			input:    []byte("Hello, World!"),
			expected: "SGVsbG8sIFdvcmxkIQ==",
		},
		{
			name:     "BinaryData",
			input:    []byte{0x00, 0x01, 0x02, 0x03, 0xff},
			expected: "AAECA/8=",
		},
		{
			name:     "UTF8Text",
			input:    []byte("中文测试"),
			expected: "5Lit5paH5rWL6K+V",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBase64(tt.input)
			assert.Equal(t, tt.expected, result, "Base64 encoding should produce expected output")
		})
	}
}

func TestFromBase64(t *testing.T) {
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
			name:      "SimpleText",
			input:     "SGVsbG8sIFdvcmxkIQ==",
			expected:  []byte("Hello, World!"),
			expectErr: false,
		},
		{
			name:      "BinaryData",
			input:     "AAECA/8=",
			expected:  []byte{0x00, 0x01, 0x02, 0x03, 0xff},
			expectErr: false,
		},
		{
			name:      "UTF8Text",
			input:     "5Lit5paH5rWL6K+V",
			expected:  []byte("中文测试"),
			expectErr: false,
		},
		{
			name:      "InvalidBase64",
			input:     "invalid!!!",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromBase64(tt.input)
			if tt.expectErr {
				assert.Error(t, err, "FromBase64 should return error for invalid input")
			} else {
				require.NoError(t, err, "FromBase64 should decode valid Base64 string without error")
				assert.Equal(t, tt.expected, result, "Decoded bytes should match expected value")
			}
		})
	}
}

func TestBase64RoundTrip(t *testing.T) {
	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err, "Random data generation should succeed")

	encoded := ToBase64(data)
	assert.NotEmpty(t, encoded, "Encoded Base64 string should not be empty")

	decoded, err := FromBase64(encoded)
	require.NoError(t, err, "Decoding Base64 string should succeed")
	assert.Equal(t, data, decoded, "Round-trip encoding/decoding should preserve original data")
}

func TestToBase64URL(t *testing.T) {
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
			name:     "DataWithSpecialChars",
			input:    []byte{0xfb, 0xff, 0xbf},
			expected: "-_-_",
		},
		{
			name:     "SimpleText",
			input:    []byte("Hello, World!"),
			expected: "SGVsbG8sIFdvcmxkIQ==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBase64URL(tt.input)
			assert.Equal(t, tt.expected, result, "Base64URL encoding should produce URL-safe output")
		})
	}
}

func TestFromBase64URL(t *testing.T) {
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
			name:      "URLSafeCharacters",
			input:     "-_-_",
			expected:  []byte{0xfb, 0xff, 0xbf},
			expectErr: false,
		},
		{
			name:      "SimpleText",
			input:     "SGVsbG8sIFdvcmxkIQ==",
			expected:  []byte("Hello, World!"),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromBase64URL(tt.input)
			if tt.expectErr {
				assert.Error(t, err, "FromBase64URL should return error for invalid input")
			} else {
				require.NoError(t, err, "FromBase64URL should decode valid Base64URL string without error")
				assert.Equal(t, tt.expected, result, "Decoded bytes should match expected value")
			}
		})
	}
}

func TestBase64UrlRoundTrip(t *testing.T) {
	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err, "Random data generation should succeed")

	encoded := ToBase64URL(data)
	assert.NotEmpty(t, encoded, "Encoded Base64URL string should not be empty")

	decoded, err := FromBase64URL(encoded)
	require.NoError(t, err, "Decoding Base64URL string should succeed")
	assert.Equal(t, data, decoded, "Round-trip encoding/decoding should preserve original data")
}

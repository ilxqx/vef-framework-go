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
			assert.Equal(t, tt.expected, result)
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
			expected:  []byte{},
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
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestBase64RoundTrip(t *testing.T) {
	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err)

	encoded := ToBase64(data)
	assert.NotEmpty(t, encoded)

	decoded, err := FromBase64(encoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

func TestToBase64Url(t *testing.T) {
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
			result := ToBase64Url(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromBase64Url(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  []byte
		expectErr bool
	}{
		{
			name:      "EmptyString",
			input:     "",
			expected:  []byte{},
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
			result, err := FromBase64Url(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestBase64UrlRoundTrip(t *testing.T) {
	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err)

	encoded := ToBase64Url(data)
	assert.NotEmpty(t, encoded)

	decoded, err := FromBase64Url(encoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

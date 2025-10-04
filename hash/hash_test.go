package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"hello", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"Hello, World!", "Hello, World!", "65a8e27d8879283831b664bd8b7f0ad4"},
		{"The quick brown fox", "The quick brown fox jumps over the lazy dog", "9e107d9d372bb6826bd81d3542a419d6"},
		{"Chinese text", "中文测试", "089b4943ea034acfa445d050c7913e55"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MD5(tc.input)
			assert.Equal(t, tc.expected, result)

			// Test MD5Bytes
			resultBytes := MD5Bytes([]byte(tc.input))
			assert.Equal(t, tc.expected, resultBytes)
		})
	}
}

func TestSHA1(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"hello", "hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"Hello, World!", "Hello, World!", "0a0a9f2a6772942557ab5355d76af442f8f65e01"},
		{"The quick brown fox", "The quick brown fox jumps over the lazy dog", "2fd4e1c67a2d28fced849ee1bb76e7391b93eb12"},
		{"Chinese text", "中文测试", "cf8a8e8f68b4e267920dba0a5f3037180cc1afd9"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SHA1(tc.input)
			assert.Equal(t, tc.expected, result)

			// Test SHA1Bytes
			resultBytes := SHA1Bytes([]byte(tc.input))
			assert.Equal(t, tc.expected, resultBytes)
		})
	}
}

func TestSHA256(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"Hello, World!", "Hello, World!", "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"},
		{"The quick brown fox", "The quick brown fox jumps over the lazy dog", "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"},
		{"Chinese text", "中文测试", "e350545d18735c5dd2dec50dcb971f3eb4cdda24b95a79bdb6b553f6a01ceb87"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SHA256(tc.input)
			assert.Equal(t, tc.expected, result)

			// Test SHA256Bytes
			resultBytes := SHA256Bytes([]byte(tc.input))
			assert.Equal(t, tc.expected, resultBytes)
		})
	}
}

func TestSHA512(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
		{"hello", "hello", "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
		{"Hello, World!", "Hello, World!", "374d794a95cdcfd8b35993185fef9ba368f160d8daf432d08ba9f1ed1e5abe6cc69291e0fa2fe0006a52570ef18c19def4e617c33ce52ef0a6e5fbe318cb0387"},
		{"Chinese text", "中文测试", "1fea9aee07bd0ab66604ef4f079d6b109a0e625c3bc38fe8f850111a9ee6b4a689f3cb454dfd8a16cbd35963382f4ca5d91cdcff2dd473028e6cfee256812eec"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SHA512(tc.input)
			assert.Equal(t, tc.expected, result)

			// Test SHA512Bytes
			resultBytes := SHA512Bytes([]byte(tc.input))
			assert.Equal(t, tc.expected, resultBytes)
		})
	}
}

func TestSM3(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "1ab21d8355cfa17f8e61194831e81a8f22bec8c728fefb747ed035eb5082aa2b"},
		{"abc", "abc", "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0"},
		{"hello", "hello", "becbbfaae6548b8bf0cfcad5a27183cd1be6093b1cceccc303d9c61d0a645268"},
		{"Hello, World!", "Hello, World!", "7ed26cbf0bee4ca7d55c1e64714c4aa7d1f163089ef5ceb603cd102c81fbcbc5"},
		{"Chinese text", "中文测试", "ac85a5ef8576c66e75c36f037aaf89bf3cbb3e2745e595bb47b47ea53f30f838"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SM3(tc.input)
			assert.Equal(t, tc.expected, result)

			// Test SM3Bytes
			resultBytes := SM3Bytes([]byte(tc.input))
			assert.Equal(t, tc.expected, resultBytes)
		})
	}
}

func TestMD5Hmac(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("test message")

	result := MD5Hmac(key, data)

	// MD5 produces 16 bytes = 32 hex chars
	assert.Len(t, result, 32)

	// Test consistency
	result2 := MD5Hmac(key, data)
	assert.Equal(t, result, result2, "MD5Hmac should produce consistent results")

	// Test different key produces different result
	differentKey := []byte("different-key")
	result3 := MD5Hmac(differentKey, data)
	assert.NotEqual(t, result, result3, "MD5Hmac with different keys should produce different results")
}

func TestSHA1Hmac(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("test message")

	result := SHA1Hmac(key, data)

	// SHA-1 produces 20 bytes = 40 hex chars
	assert.Len(t, result, 40)
}

func TestSHA256Hmac(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("test message")

	result := SHA256Hmac(key, data)

	// SHA-256 produces 32 bytes = 64 hex chars
	assert.Len(t, result, 64)

	// Test with known values
	testCases := []struct {
		name     string
		key      string
		data     string
		expected string
	}{
		{
			name:     "standard test vector",
			key:      "key",
			data:     "The quick brown fox jumps over the lazy dog",
			expected: "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SHA256Hmac([]byte(tc.key), []byte(tc.data))
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSHA512Hmac(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("test message")

	result := SHA512Hmac(key, data)

	// SHA-512 produces 64 bytes = 128 hex chars
	assert.Len(t, result, 128)
}

func TestSM3Hmac(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("test message")

	result := SM3Hmac(key, data)

	// SM3 produces 32 bytes = 64 hex chars
	assert.Len(t, result, 64)

	// Test consistency
	result2 := SM3Hmac(key, data)
	assert.Equal(t, result, result2, "SM3Hmac should produce consistent results")
}

func TestHashFunctions_NilInput(t *testing.T) {
	// Test that hash functions handle nil bytes gracefully
	t.Run("MD5Bytes with nil", func(t *testing.T) {
		result := MD5Bytes(nil)
		assert.NotEmpty(t, result)
	})

	t.Run("SHA256Bytes with nil", func(t *testing.T) {
		result := SHA256Bytes(nil)
		assert.NotEmpty(t, result)
	})

	t.Run("SM3Bytes with nil", func(t *testing.T) {
		result := SM3Bytes(nil)
		assert.NotEmpty(t, result)
	})
}

func TestHmacFunctions_EmptyKeyOrData(t *testing.T) {
	t.Run("SHA256Hmac with empty key", func(t *testing.T) {
		result := SHA256Hmac([]byte{}, []byte("data"))
		assert.Len(t, result, 64)
	})

	t.Run("SHA256Hmac with empty data", func(t *testing.T) {
		result := SHA256Hmac([]byte("key"), []byte{})
		assert.Len(t, result, 64)
	})

	t.Run("SHA256Hmac with both empty", func(t *testing.T) {
		result := SHA256Hmac([]byte{}, []byte{})
		assert.Len(t, result, 64)
	})
}

func TestHashOutputFormat(t *testing.T) {
	// Test that all hash functions return lowercase hex strings
	testData := "test"

	t.Run("MD5 output is lowercase hex", func(t *testing.T) {
		result := MD5(testData)
		assert.Regexp(t, "^[0-9a-f]+$", result)
	})

	t.Run("SHA256 output is lowercase hex", func(t *testing.T) {
		result := SHA256(testData)
		assert.Regexp(t, "^[0-9a-f]+$", result)
	})

	t.Run("SM3 output is lowercase hex", func(t *testing.T) {
		result := SM3(testData)
		assert.Regexp(t, "^[0-9a-f]+$", result)
	})
}

func BenchmarkMD5(b *testing.B) {
	data := "benchmark test data"

	for b.Loop() {
		MD5(data)
	}
}

func BenchmarkSHA256(b *testing.B) {
	data := "benchmark test data"

	for b.Loop() {
		SHA256(data)
	}
}

func BenchmarkSHA512(b *testing.B) {
	data := "benchmark test data"

	for b.Loop() {
		SHA512(data)
	}
}

func BenchmarkSM3(b *testing.B) {
	data := "benchmark test data"

	for b.Loop() {
		SM3(data)
	}
}

func BenchmarkSHA256Hmac(b *testing.B) {
	key := []byte("secret-key")
	data := []byte("benchmark test data")

	for b.Loop() {
		SHA256Hmac(key, data)
	}
}

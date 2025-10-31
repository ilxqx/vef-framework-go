package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAESCipher_CBC tests AES encryption and decryption in CBC mode.
func TestAESCipher_CBC(t *testing.T) {
	key := make([]byte, 32)
	iv := make([]byte, 16)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	cipher, err := NewAES(key, iv, AesModeCBC)
	require.NoError(t, err, "Should create AES cipher in CBC mode")

	tests := []struct {
		name      string
		plaintext string
	}{
		{"EnglishText", "Hello, World!"},
		{"LongerMessage", "This is a test message"},
		{"WithDescription", "AES-256-CBC encryption test"},
		{"ChineseCharacters", "中文测试"},
		{"SpecialCharacters", "Special chars: !@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := cipher.Encrypt(tt.plaintext)
			require.NoError(t, err, "Should encrypt plaintext successfully")

			decrypted, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err, "Should decrypt ciphertext successfully")

			assert.Equal(t, tt.plaintext, decrypted, "Decrypted text should match original plaintext")
		})
	}
}

// TestAESCipher_GCM tests AES encryption and decryption in GCM mode.
func TestAESCipher_GCM(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := NewAES(key, nil, AesModeGCM)
	require.NoError(t, err, "Should create AES cipher in GCM mode")

	tests := []struct {
		name      string
		plaintext string
	}{
		{"EnglishText", "Hello, World!"},
		{"LongerMessage", "This is a test message"},
		{"WithDescription", "AES-256-GCM encryption test"},
		{"ChineseCharacters", "中文测试"},
		{"SpecialCharacters", "Special chars: !@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := cipher.Encrypt(tt.plaintext)
			require.NoError(t, err, "Should encrypt plaintext successfully")

			decrypted, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err, "Should decrypt ciphertext successfully")

			assert.Equal(t, tt.plaintext, decrypted, "Decrypted text should match original plaintext")
		})
	}
}

// TestAESCipher_FromHex tests creating AES cipher from hex-encoded key and IV.
func TestAESCipher_FromHex(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	ivHex := "0123456789abcdef0123456789abcdef"

	cipher, err := NewAESFromHex(keyHex, ivHex, AesModeCBC)
	require.NoError(t, err, "Should create AES cipher from hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestAESCipher_FromBase64 tests creating AES cipher from base64-encoded key and IV.
func TestAESCipher_FromBase64(t *testing.T) {
	key := make([]byte, 32)
	iv := make([]byte, 16)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	keyBase64 := base64.StdEncoding.EncodeToString(key)
	ivBase64 := base64.StdEncoding.EncodeToString(iv)

	cipher, err := NewAESFromBase64(keyBase64, ivBase64, AesModeCBC)
	require.NoError(t, err, "Should create AES cipher from base64")

	plaintext := "Test message with base64 encoded key"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestAESCipher_InvalidKeySize tests that invalid key size is rejected.
func TestAESCipher_InvalidKeySize(t *testing.T) {
	invalidKey := make([]byte, 15)
	iv := make([]byte, 16)

	_, err := NewAES(invalidKey, iv, AesModeCBC)
	assert.Error(t, err, "Should reject invalid key size")
}

// TestAESCipher_InvalidIVSize tests that invalid IV size is rejected.
func TestAESCipher_InvalidIVSize(t *testing.T) {
	key := make([]byte, 32)
	invalidIV := make([]byte, 8)

	_, err := NewAES(key, invalidIV, AesModeCBC)
	assert.Error(t, err, "Should reject invalid IV size")
}

// TestAESCipher_GCM_Authentication tests that GCM mode detects tampering.
func TestAESCipher_GCM_Authentication(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := NewAES(key, nil, AesModeGCM)
	require.NoError(t, err, "Should create AES cipher in GCM mode")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	tamperedCiphertext := ciphertext[:len(ciphertext)-2] + "X" + ciphertext[len(ciphertext)-1:]

	_, err = cipher.Decrypt(tamperedCiphertext)
	assert.Error(t, err, "Should reject tampered ciphertext")
}

// TestAESCipher_KeySizes tests AES with different key sizes (128, 192, 256 bits).
func TestAESCipher_KeySizes(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
	}{
		{"AES-128", 16},
		{"AES-192", 24},
		{"AES-256", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			iv := make([]byte, 16)
			_, err := rand.Read(key)
			require.NoError(t, err)
			_, err = rand.Read(iv)
			require.NoError(t, err)

			cipher, err := NewAES(key, iv, AesModeCBC)
			require.NoError(t, err, "Should create AES cipher with %d-byte key", tt.keySize)

			plaintext := "Test message"
			ciphertext, err := cipher.Encrypt(plaintext)
			require.NoError(t, err, "Should encrypt plaintext successfully")

			decrypted, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err, "Should decrypt ciphertext successfully")

			assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
		})
	}
}

// TestAESCipher_DefaultMode tests that default mode is GCM.
func TestAESCipher_DefaultMode(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := NewAES(key, nil)
	require.NoError(t, err, "Should create AES cipher with default mode")

	plaintext := "Test default mode"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")

	aesCipher, ok := cipher.(*AesCipher)
	require.True(t, ok)
	assert.Equal(t, AesModeGCM, aesCipher.mode, "Default mode should be GCM")
}

// TestPKCS7Padding tests PKCS7 padding and unpadding.
func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		blockSize int
	}{
		{"ShortString", "Hello", 16},
		{"EightByteBlock", "Test", 8},
		{"ExactBlock", "1234567890123456", 16},
		{"EmptyString", "", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := pkcs7Padding([]byte(tt.input), tt.blockSize)

			assert.Equal(t, 0, len(padded)%tt.blockSize,
				"Padded length should be multiple of block size")

			unpadded, err := pkcs7Unpadding(padded)
			require.NoError(t, err, "Should unpad successfully")

			assert.Equal(t, tt.input, string(unpadded),
				"Unpadded text should match original input")
		})
	}
}

package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESCipher_CBC(t *testing.T) {
	// Generate random key and IV
	key := make([]byte, 32) // AES-256
	iv := make([]byte, 16)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	cipher, err := NewAES(key, iv, AESModeCBC)
	require.NoError(t, err, "Failed to create AES cipher")

	testCases := []string{
		"Hello, World!",
		"This is a test message",
		"AES-256-CBC encryption test",
		"中文测试",
		"Special chars: !@#$%^&*()",
	}

	for _, plaintext := range testCases {
		t.Run(plaintext, func(t *testing.T) {
			// Encrypt
			ciphertext, err := cipher.Encrypt(plaintext)
			require.NoError(t, err, "Encryption failed")

			// Decrypt
			decrypted, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err, "Decryption failed")

			// Verify
			assert.Equal(t, plaintext, decrypted)
		})
	}
}

func TestAESCipher_GCM(t *testing.T) {
	// Generate random key
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := NewAES(key, nil, AESModeGCM)
	require.NoError(t, err, "Failed to create AES cipher")

	testCases := []string{
		"Hello, World!",
		"This is a test message",
		"AES-256-GCM encryption test",
		"中文测试",
		"Special chars: !@#$%^&*()",
	}

	for _, plaintext := range testCases {
		t.Run(plaintext, func(t *testing.T) {
			// Encrypt
			ciphertext, err := cipher.Encrypt(plaintext)
			require.NoError(t, err, "Encryption failed")

			// Decrypt
			decrypted, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err, "Decryption failed")

			// Verify
			assert.Equal(t, plaintext, decrypted)
		})
	}
}

func TestAESCipher_FromHex(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 32 bytes
	ivHex := "0123456789abcdef0123456789abcdef"                                  // 16 bytes

	cipher, err := NewAESFromHex(keyHex, ivHex, AESModeCBC)
	require.NoError(t, err, "Failed to create AES cipher from hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestAESCipher_FromBase64(t *testing.T) {
	// Generate random key and IV
	key := make([]byte, 32)
	iv := make([]byte, 16)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	keyBase64 := base64.StdEncoding.EncodeToString(key)
	ivBase64 := base64.StdEncoding.EncodeToString(iv)

	cipher, err := NewAESFromBase64(keyBase64, ivBase64, AESModeCBC)
	require.NoError(t, err, "Failed to create AES cipher from base64")

	plaintext := "Test message with base64 encoded key"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestAESCipher_InvalidKeySize(t *testing.T) {
	invalidKey := make([]byte, 15) // Invalid key size
	iv := make([]byte, 16)

	_, err := NewAES(invalidKey, iv, AESModeCBC)
	assert.Error(t, err, "Expected error for invalid key size")
}

func TestAESCipher_InvalidIVSize(t *testing.T) {
	key := make([]byte, 32)
	invalidIV := make([]byte, 8) // Invalid IV size

	_, err := NewAES(key, invalidIV, AESModeCBC)
	assert.Error(t, err, "Expected error for invalid IV size")
}

func TestAESCipher_GCM_Authentication(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := NewAES(key, nil, AESModeGCM)
	require.NoError(t, err, "Failed to create AES cipher")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	// Tamper with ciphertext (change one byte)
	tamperedCiphertext := ciphertext[:len(ciphertext)-2] + "X" + ciphertext[len(ciphertext)-1:]

	_, err = cipher.Decrypt(tamperedCiphertext)
	assert.Error(t, err, "Expected error for tampered ciphertext")
}

func TestAESCipher_KeySizes(t *testing.T) {
	keySizes := []int{16, 24, 32} // AES-128, AES-192, AES-256

	for _, size := range keySizes {
		t.Run(fmt.Sprintf("AES-%d", size*8), func(t *testing.T) {
			key := make([]byte, size)
			iv := make([]byte, 16)
			_, err := rand.Read(key)
			require.NoError(t, err)
			_, err = rand.Read(iv)
			require.NoError(t, err)

			cipher, err := NewAES(key, iv, AESModeCBC)
			require.NoError(t, err, "Failed to create AES cipher with %d-byte key", size)

			plaintext := "Test message"
			ciphertext, err := cipher.Encrypt(plaintext)
			require.NoError(t, err, "Encryption failed")

			decrypted, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err, "Decryption failed")

			assert.Equal(t, plaintext, decrypted)
		})
	}
}

func TestAESCipher_DefaultMode(t *testing.T) {
	// Test that default mode is GCM
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := NewAES(key, nil)
	require.NoError(t, err, "Failed to create AES cipher with default mode")

	plaintext := "Test default mode"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)

	// Verify it's using GCM by checking that the cipher is an AESCipher with GCM mode
	aesCipher, ok := cipher.(*AESCipher)
	require.True(t, ok)
	assert.Equal(t, AESModeGCM, aesCipher.mode)
}

func TestPKCS7Padding(t *testing.T) {
	testCases := []struct {
		input     string
		blockSize int
	}{
		{"Hello", 16},
		{"Test", 8},
		{"1234567890123456", 16}, // Exactly one block
		{"", 16},                 // Empty input
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			padded := pkcs7Padding([]byte(tc.input), tc.blockSize)

			// Verify padding length
			assert.Equal(t, 0, len(padded)%tc.blockSize, "Padded length %d is not a multiple of block size %d", len(padded), tc.blockSize)

			// Unpad
			unpadded, err := pkcs7Unpadding(padded)
			require.NoError(t, err, "Failed to unpad")

			// Verify
			assert.Equal(t, tc.input, string(unpadded))
		})
	}
}

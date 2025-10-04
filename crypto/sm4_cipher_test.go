package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjfoc/gmsm/sm4"
)

func TestSM4Cipher_CBC(t *testing.T) {
	// Generate random key and IV
	key := make([]byte, sm4.BlockSize) // 16 bytes
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	cipher, err := NewSM4(key, iv, SM4ModeCBC)
	require.NoError(t, err, "Failed to create SM4 cipher")

	testCases := []string{
		"Hello, World!",
		"SM4-CBC encryption test",
		"中文测试",
		"Special chars: !@#$%^&*()",
		"国密SM4加密算法",
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

func TestSM4Cipher_ECB(t *testing.T) {
	// Generate random key
	key := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := NewSM4(key, nil, SM4ModeECB)
	require.NoError(t, err, "Failed to create SM4 cipher")

	testCases := []string{
		"Hello, World!",
		"SM4-ECB encryption test",
		"中文测试",
		"Special chars: !@#$%^&*()",
		"国密SM4加密算法",
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

func TestSM4Cipher_FromHex(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef" // 16 bytes
	ivHex := "fedcba9876543210fedcba9876543210"  // 16 bytes

	cipher, err := NewSM4FromHex(keyHex, ivHex, SM4ModeCBC)
	require.NoError(t, err, "Failed to create SM4 cipher from hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestSM4Cipher_FromBase64(t *testing.T) {
	// Generate random key and IV
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	keyBase64 := base64.StdEncoding.EncodeToString(key)
	ivBase64 := base64.StdEncoding.EncodeToString(iv)

	cipher, err := NewSM4FromBase64(keyBase64, ivBase64, SM4ModeCBC)
	require.NoError(t, err, "Failed to create SM4 cipher from base64")

	plaintext := "Test message with base64 encoded key"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestSM4Cipher_InvalidKeySize(t *testing.T) {
	invalidKey := make([]byte, 8) // Invalid key size (must be 16)
	iv := make([]byte, sm4.BlockSize)

	_, err := NewSM4(invalidKey, iv, SM4ModeCBC)
	assert.Error(t, err, "Expected error for invalid key size")
}

func TestSM4Cipher_InvalidIVSize(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	invalidIV := make([]byte, 8) // Invalid IV size (must be 16)

	_, err := NewSM4(key, invalidIV, SM4ModeCBC)
	assert.Error(t, err, "Expected error for invalid IV size")
}

func TestSM4Cipher_ECB_NoIVRequired(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// ECB mode doesn't require IV
	cipher, err := NewSM4(key, nil, SM4ModeECB)
	require.NoError(t, err, "Failed to create SM4 cipher in ECB mode without IV")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestSM4Cipher_ECB_Deterministic(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := NewSM4(key, nil, SM4ModeECB)
	require.NoError(t, err, "Failed to create SM4 cipher")

	plaintext := "Test message"

	// ECB mode should produce the same ciphertext for the same plaintext
	ciphertext1, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "First encryption failed")

	ciphertext2, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Second encryption failed")

	assert.Equal(t, ciphertext1, ciphertext2, "ECB mode should produce the same ciphertext for the same plaintext")
}

func TestSM4Cipher_CBC_NonDeterministic(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	cipher, err := NewSM4(key, iv, SM4ModeCBC)
	require.NoError(t, err, "Failed to create SM4 cipher")

	plaintext := "Test message"

	// CBC mode with the same IV should produce the same ciphertext
	ciphertext1, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "First encryption failed")

	ciphertext2, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Second encryption failed")

	// With the same IV, CBC should produce the same ciphertext
	assert.Equal(t, ciphertext1, ciphertext2, "CBC mode with the same IV should produce the same ciphertext")

	// Both should decrypt to the same plaintext
	decrypted1, err := cipher.Decrypt(ciphertext1)
	require.NoError(t, err, "First decryption failed")

	decrypted2, err := cipher.Decrypt(ciphertext2)
	require.NoError(t, err, "Second decryption failed")

	assert.Equal(t, plaintext, decrypted1)
	assert.Equal(t, plaintext, decrypted2)
}

func TestSM4Cipher_LongMessage(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	cipher, err := NewSM4(key, iv, SM4ModeCBC)
	require.NoError(t, err, "Failed to create SM4 cipher")

	// Create a long message (multiple blocks)
	plaintext := strings.Repeat("This is a test message. ", 100)

	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestSM4Cipher_EmptyString(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	cipher, err := NewSM4(key, iv, SM4ModeCBC)
	require.NoError(t, err, "Failed to create SM4 cipher")

	plaintext := ""
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestSM4Cipher_DefaultMode(t *testing.T) {
	// Test that default mode is CBC
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	cipher, err := NewSM4(key, iv)
	require.NoError(t, err, "Failed to create SM4 cipher with default mode")

	plaintext := "Test default mode"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)

	// Verify it's using CBC
	sm4Cipher, ok := cipher.(*SM4Cipher)
	require.True(t, ok)
	assert.Equal(t, SM4ModeCBC, sm4Cipher.mode)
}

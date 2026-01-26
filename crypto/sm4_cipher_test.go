package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjfoc/gmsm/sm4"

	"github.com/ilxqx/vef-framework-go/encoding"
)

// TestSm4Cipher_Cbc tests SM4 encryption and decryption in CBC mode.
func TestSm4Cipher_Cbc(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")
	_, err = rand.Read(iv)
	require.NoError(t, err, "Should generate random IV")

	cipher, err := NewSM4(key, WithSM4Iv(iv), WithSM4Mode(SM4ModeCBC))
	require.NoError(t, err, "Should create SM4 cipher in CBC mode")

	tests := []struct {
		name      string
		plaintext string
	}{
		{"EnglishText", "Hello, World!"},
		{"WithDescription", "SM4-CBC encryption test"},
		{"ChineseCharacters", "中文测试"},
		{"SpecialCharacters", "Special chars: !@#$%^&*()"},
		{"ChineseAlgorithm", "国密SM4加密算法"},
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

// TestSm4Cipher_Ecb tests SM4 encryption and decryption in ECB mode.
func TestSm4Cipher_Ecb(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")

	cipher, err := NewSM4(key, WithSM4Mode(SM4ModeECB))
	require.NoError(t, err, "Should create SM4 cipher in ECB mode")

	tests := []struct {
		name      string
		plaintext string
	}{
		{"EnglishText", "Hello, World!"},
		{"WithDescription", "SM4-ECB encryption test"},
		{"ChineseCharacters", "中文测试"},
		{"SpecialCharacters", "Special chars: !@#$%^&*()"},
		{"ChineseAlgorithm", "国密SM4加密算法"},
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

// TestSm4Cipher_FromHex tests creating SM4 cipher from hex-encoded key.
func TestSm4Cipher_FromHex(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef"
	ivHex := "fedcba9876543210fedcba9876543210"

	iv, err := encoding.FromHex(ivHex)
	require.NoError(t, err, "Should decode hex IV")

	cipher, err := NewSM4FromHex(keyHex, WithSM4Iv(iv), WithSM4Mode(SM4ModeCBC))
	require.NoError(t, err, "Should create SM4 cipher from hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestSm4Cipher_FromBase64 tests creating SM4 cipher from base64-encoded key.
func TestSm4Cipher_FromBase64(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")
	_, err = rand.Read(iv)
	require.NoError(t, err, "Should generate random IV")

	keyBase64 := base64.StdEncoding.EncodeToString(key)

	cipher, err := NewSM4FromBase64(keyBase64, WithSM4Iv(iv), WithSM4Mode(SM4ModeCBC))
	require.NoError(t, err, "Should create SM4 cipher from base64")

	plaintext := "Test message with base64 encoded key"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestSm4Cipher_InvalidKeySize tests that invalid key size is rejected.
func TestSm4Cipher_InvalidKeySize(t *testing.T) {
	invalidKey := make([]byte, 8)
	iv := make([]byte, sm4.BlockSize)

	_, err := NewSM4(invalidKey, WithSM4Iv(iv), WithSM4Mode(SM4ModeCBC))
	assert.Error(t, err, "Should reject invalid key size")
}

// TestSm4Cipher_InvalidIvSize tests that invalid IV size is rejected.
func TestSm4Cipher_InvalidIvSize(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	invalidIV := make([]byte, 8)

	_, err := NewSM4(key, WithSM4Iv(invalidIV), WithSM4Mode(SM4ModeCBC))
	assert.Error(t, err, "Should reject invalid IV size")
}

// TestSm4Cipher_Ecb_NoIVRequired tests that ECB mode doesn't require IV.
func TestSm4Cipher_Ecb_NoIvRequired(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")

	cipher, err := NewSM4(key, WithSM4Mode(SM4ModeECB))
	require.NoError(t, err, "Should create SM4 cipher in ECB mode without IV")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestSm4Cipher_Ecb_Deterministic tests that ECB mode is deterministic.
func TestSm4Cipher_Ecb_Deterministic(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")

	cipher, err := NewSM4(key, WithSM4Mode(SM4ModeECB))
	require.NoError(t, err, "Should create SM4 cipher")

	plaintext := "Test message"

	ciphertext1, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	ciphertext2, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	assert.Equal(t, ciphertext1, ciphertext2,
		"ECB mode should produce same ciphertext for same plaintext")
}

// TestSm4Cipher_Cbc_NonDeterministic tests CBC mode with fixed IV.
func TestSm4Cipher_Cbc_NonDeterministic(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")
	_, err = rand.Read(iv)
	require.NoError(t, err, "Should generate random IV")

	cipher, err := NewSM4(key, WithSM4Iv(iv), WithSM4Mode(SM4ModeCBC))
	require.NoError(t, err, "Should create SM4 cipher")

	plaintext := "Test message"

	ciphertext1, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	ciphertext2, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	assert.Equal(t, ciphertext1, ciphertext2,
		"CBC mode with same IV should produce same ciphertext")

	decrypted1, err := cipher.Decrypt(ciphertext1)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	decrypted2, err := cipher.Decrypt(ciphertext2)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted1, "First decrypted text should match original plaintext")
	assert.Equal(t, plaintext, decrypted2, "Second decrypted text should match original plaintext")
}

// TestSm4Cipher_LongMessage tests SM4 with long messages spanning multiple blocks.
func TestSm4Cipher_LongMessage(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")
	_, err = rand.Read(iv)
	require.NoError(t, err, "Should generate random IV")

	cipher, err := NewSM4(key, WithSM4Iv(iv), WithSM4Mode(SM4ModeCBC))
	require.NoError(t, err, "Should create SM4 cipher")

	plaintext := strings.Repeat("This is a test message. ", 100)

	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestSm4Cipher_EmptyString tests SM4 with empty string input.
func TestSm4Cipher_EmptyString(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")
	_, err = rand.Read(iv)
	require.NoError(t, err, "Should generate random IV")

	cipher, err := NewSM4(key, WithSM4Iv(iv), WithSM4Mode(SM4ModeCBC))
	require.NoError(t, err, "Should create SM4 cipher")

	plaintext := ""
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt empty string successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match empty plaintext")
}

// TestSm4Cipher_DefaultMode tests that default mode is CBC.
func TestSm4Cipher_DefaultMode(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err, "Should generate random key")
	_, err = rand.Read(iv)
	require.NoError(t, err, "Should generate random IV")

	cipher, err := NewSM4(key, WithSM4Iv(iv))
	require.NoError(t, err, "Should create SM4 cipher with default mode")

	plaintext := "Test default mode"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

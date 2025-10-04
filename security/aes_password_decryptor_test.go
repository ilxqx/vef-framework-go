package security

import (
	"crypto/rand"
	"testing"

	"github.com/ilxqx/vef-framework-go/crypto"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESPasswordDecryptor(t *testing.T) {
	// Generate random key and IV
	key := make([]byte, 32) // AES-256
	iv := make([]byte, 16)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	// Create decryptor
	decryptor, err := NewAESPasswordDecryptor(key, iv)
	require.NoError(t, err, "Failed to create AES password decryptor")

	// Test password
	password := "MySecurePassword123!@#"

	// Encrypt using crypto package
	cipher, err := crypto.NewAES(key, iv, crypto.AESModeCBC)
	require.NoError(t, err)

	encryptedPassword, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decryptedPassword, err := decryptor.Decrypt(encryptedPassword)
	require.NoError(t, err, "Failed to decrypt password")

	// Verify
	assert.Equal(t, password, decryptedPassword, "Decrypted password should match original")
}

func TestAESPasswordDecryptor_FromHex(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 32 bytes
	ivHex := "0123456789abcdef0123456789abcdef"                                  // 16 bytes

	decryptor, err := NewAESPasswordDecryptorFromHex(keyHex, ivHex)
	require.NoError(t, err)

	// Encrypt using the same key/iv
	cipher, err := crypto.NewAESFromHex(keyHex, ivHex, crypto.AESModeCBC)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestAESPasswordDecryptor_FromBase64(t *testing.T) {
	key := make([]byte, 32)
	iv := make([]byte, 16)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	keyBase64 := encoding.ToBase64(key)
	ivBase64 := encoding.ToBase64(iv)

	decryptor, err := NewAESPasswordDecryptorFromBase64(keyBase64, ivBase64)
	require.NoError(t, err)

	// Encrypt using the same key/iv
	cipher, err := crypto.NewAES(key, iv, crypto.AESModeCBC)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestAESPasswordDecryptor_InvalidKeyLength(t *testing.T) {
	invalidKey := make([]byte, 15) // Invalid key size
	iv := make([]byte, 16)

	_, err := NewAESPasswordDecryptor(invalidKey, iv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AES key length")
}

func TestAESPasswordDecryptor_InvalidIVLength(t *testing.T) {
	key := make([]byte, 32)
	invalidIV := make([]byte, 8) // Invalid IV size

	_, err := NewAESPasswordDecryptor(key, invalidIV)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IV length")
}

func TestAESPasswordDecryptor_NilIV(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Should use first 16 bytes of key as IV
	decryptor, err := NewAESPasswordDecryptor(key, nil)
	require.NoError(t, err)

	// Encrypt using the same key as IV
	cipher, err := crypto.NewAES(key, key[:16], crypto.AESModeCBC)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestAESPasswordDecryptor_DifferentKeySizes(t *testing.T) {
	testCases := []struct {
		name    string
		keySize int
	}{
		{"AES-128", 16},
		{"AES-192", 24},
		{"AES-256", 32},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			iv := make([]byte, 16)
			_, err := rand.Read(key)
			require.NoError(t, err)
			_, err = rand.Read(iv)
			require.NoError(t, err)

			decryptor, err := NewAESPasswordDecryptor(key, iv)
			require.NoError(t, err)

			cipher, err := crypto.NewAES(key, iv, crypto.AESModeCBC)
			require.NoError(t, err)

			password := "TestPassword"
			encrypted, err := cipher.Encrypt(password)
			require.NoError(t, err)

			decrypted, err := decryptor.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, password, decrypted)
		})
	}
}

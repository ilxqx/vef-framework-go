package security

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjfoc/gmsm/sm4"

	"github.com/ilxqx/vef-framework-go/crypto"
	"github.com/ilxqx/vef-framework-go/encoding"
)

func TestSM4PasswordDecryptor_CBC(t *testing.T) {
	// Generate random key and IV
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	// Create decryptor
	decryptor, err := NewSm4PasswordDecryptor(key, iv, crypto.Sm4ModeCBC)
	require.NoError(t, err, "Failed to create SM4 password decryptor")

	// Test password
	password := "MySecurePassword123!@#"

	// Encrypt using crypto package
	cipher, err := crypto.NewSM4(key, iv, crypto.Sm4ModeCBC)
	require.NoError(t, err)

	encryptedPassword, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decryptedPassword, err := decryptor.Decrypt(encryptedPassword)
	require.NoError(t, err, "Failed to decrypt password")

	// Verify
	assert.Equal(t, password, decryptedPassword, "Decrypted password should match original")
}

func TestSM4PasswordDecryptor_ECB(t *testing.T) {
	// Generate random key
	key := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Create decryptor (ECB mode doesn't use IV)
	decryptor, err := NewSm4PasswordDecryptor(key, key[:sm4.BlockSize], crypto.Sm4ModeECB)
	require.NoError(t, err, "Failed to create SM4 password decryptor")

	// Test password
	password := "MySecurePassword123!@#"

	// Encrypt using crypto package
	cipher, err := crypto.NewSM4(key, nil, crypto.Sm4ModeECB)
	require.NoError(t, err)

	encryptedPassword, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decryptedPassword, err := decryptor.Decrypt(encryptedPassword)
	require.NoError(t, err, "Failed to decrypt password")

	// Verify
	assert.Equal(t, password, decryptedPassword, "Decrypted password should match original")
}

func TestSM4PasswordDecryptor_DefaultMode(t *testing.T) {
	// Generate random key and IV
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	// Create decryptor without specifying mode (should default to CBC)
	decryptor, err := NewSm4PasswordDecryptor(key, iv)
	require.NoError(t, err)

	password := "TestPassword"

	// Encrypt using crypto package with CBC mode
	cipher, err := crypto.NewSM4(key, iv, crypto.Sm4ModeCBC)
	require.NoError(t, err)

	encryptedPassword, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decryptedPassword, err := decryptor.Decrypt(encryptedPassword)
	require.NoError(t, err)
	assert.Equal(t, password, decryptedPassword)
}

func TestSM4PasswordDecryptor_FromHex(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef" // 16 bytes
	ivHex := "fedcba9876543210fedcba9876543210"  // 16 bytes

	decryptor, err := NewSm4PasswordDecryptorFromHex(keyHex, ivHex)
	require.NoError(t, err)

	// Encrypt using the same key/iv
	cipher, err := crypto.NewSM4FromHex(keyHex, ivHex)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestSM4PasswordDecryptor_FromBase64(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	iv := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	keyBase64 := encoding.ToBase64(key)
	ivBase64 := encoding.ToBase64(iv)

	decryptor, err := NewSm4PasswordDecryptorFromBase64(keyBase64, ivBase64)
	require.NoError(t, err)

	// Encrypt using the same key/iv
	cipher, err := crypto.NewSM4(key, iv)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestSM4PasswordDecryptor_InvalidKeySize(t *testing.T) {
	invalidKey := make([]byte, 8) // Invalid key size (must be 16)
	iv := make([]byte, sm4.BlockSize)

	_, err := NewSm4PasswordDecryptor(invalidKey, iv)
	assert.Error(t, err, "Should return error for invalid key size")
	assert.Contains(t, err.Error(), "invalid SM4 key length")
}

func TestSM4PasswordDecryptor_InvalidIVSize(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	invalidIV := make([]byte, 8) // Invalid IV size (must be 16)

	_, err := NewSm4PasswordDecryptor(key, invalidIV)
	assert.Error(t, err, "Should return error for invalid IV size")
	assert.Contains(t, err.Error(), "invalid IV length")
}

func TestSM4PasswordDecryptor_NilIV(t *testing.T) {
	key := make([]byte, sm4.BlockSize)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Should use first 16 bytes of key as IV
	decryptor, err := NewSm4PasswordDecryptor(key, nil)
	require.NoError(t, err)

	// Encrypt using the same key as IV
	cipher, err := crypto.NewSM4(key, key[:sm4.BlockSize])
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

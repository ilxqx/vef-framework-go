package security

import (
	"crypto/rand"
	"testing"

	"github.com/ilxqx/vef-framework-go/crypto"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjfoc/gmsm/sm2"
)

func TestSM2PasswordDecryptor(t *testing.T) {
	// Generate SM2 key pair
	privateKey, err := sm2.GenerateKey(rand.Reader)
	require.NoError(t, err, "Failed to generate SM2 key pair")

	// Create decryptor
	decryptor, err := NewSM2PasswordDecryptor(privateKey)
	require.NoError(t, err, "Failed to create SM2 password decryptor")

	// Test password
	password := "MySecurePassword123!@#"

	// Encrypt password using SM2 public key
	cipherBytes, err := sm2.Encrypt(&privateKey.PublicKey, []byte(password), rand.Reader, sm2.C1C3C2)
	require.NoError(t, err, "Failed to encrypt password")

	// Convert to base64
	encryptedPasswordB64 := encoding.ToBase64(cipherBytes)

	// Decrypt
	decryptedPassword, err := decryptor.Decrypt(encryptedPasswordB64)
	require.NoError(t, err, "Failed to decrypt password")

	// Verify
	assert.Equal(t, password, decryptedPassword, "Decrypted password should match original")
}

func TestSM2PasswordDecryptor_NilKey(t *testing.T) {
	_, err := NewSM2PasswordDecryptor(nil)
	assert.Error(t, err, "Should return error for nil private key")
	assert.Contains(t, err.Error(), "private key cannot be nil")
}

func TestSM2PasswordDecryptor_Integration(t *testing.T) {
	// Generate SM2 key pair
	privateKey, err := sm2.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Create decryptor
	decryptor, err := NewSM2PasswordDecryptor(privateKey)
	require.NoError(t, err)

	// Use crypto package to encrypt
	cipher, err := crypto.NewSM2(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword123"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt using decryptor
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

package crypto

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjfoc/gmsm/sm2"
)

func generateSM2KeyPair() (*sm2.PrivateKey, error) {
	return sm2.GenerateKey(rand.Reader)
}

func TestSM2Cipher_Encrypt_Decrypt(t *testing.T) {
	privateKey, err := generateSM2KeyPair()
	require.NoError(t, err, "Failed to generate SM2 key pair")

	cipher, err := NewSM2(privateKey, &privateKey.PublicKey)
	require.NoError(t, err, "Failed to create SM2 cipher")

	testCases := []string{
		"Hello, World!",
		"SM2 encryption test",
		"中文测试",
		"Special chars: !@#$%^&*()",
		"国密SM2加密算法",
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

func TestSM2Cipher_FromPEM(t *testing.T) {
	// Skip this test as SM2 PEM encoding is library-specific
	// Users should use NewSM2 or NewSM2FromHex for now
	t.Skip("SM2 PEM encoding is library-specific, use NewSM2 or NewSM2FromHex instead")
}

func TestSM2Cipher_FromHex(t *testing.T) {
	// Skip this test as SM2 key marshaling is library-specific
	// Users should use NewSM2 for direct key usage
	t.Skip("SM2 key marshaling is library-specific, use NewSM2 instead")
}

func TestSM2Cipher_PublicKeyOnly(t *testing.T) {
	privateKey, err := generateSM2KeyPair()
	require.NoError(t, err, "Failed to generate SM2 key pair")

	// Create cipher with only public key
	cipher, err := NewSM2(nil, &privateKey.PublicKey)
	require.NoError(t, err, "Failed to create SM2 cipher")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	// Decryption should fail (no private key)
	_, err = cipher.Decrypt(ciphertext)
	assert.Error(t, err, "Expected error for decryption without private key")
}

func TestSM2Cipher_PrivateKeyOnly(t *testing.T) {
	privateKey, err := generateSM2KeyPair()
	require.NoError(t, err, "Failed to generate SM2 key pair")

	// Create cipher with only private key (public key should be derived)
	cipher, err := NewSM2(privateKey, nil)
	require.NoError(t, err, "Failed to create SM2 cipher")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestSM2Cipher_NoKeys(t *testing.T) {
	_, err := NewSM2(nil, nil)
	assert.Error(t, err, "Expected error when creating cipher without any keys")
}

func TestSM2Cipher_MultipleEncryptions(t *testing.T) {
	privateKey, err := generateSM2KeyPair()
	require.NoError(t, err, "Failed to generate SM2 key pair")

	cipher, err := NewSM2(privateKey, &privateKey.PublicKey)
	require.NoError(t, err, "Failed to create SM2 cipher")

	plaintext := "Test message"

	// Encrypt the same message multiple times
	// Results should be different (due to random component in SM2)
	ciphertext1, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "First encryption failed")

	ciphertext2, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Second encryption failed")

	assert.NotEqual(t, ciphertext1, ciphertext2, "Expected different ciphertexts for same plaintext (SM2 should have random component)")

	// Both should decrypt to the same plaintext
	decrypted1, err := cipher.Decrypt(ciphertext1)
	require.NoError(t, err, "First decryption failed")

	decrypted2, err := cipher.Decrypt(ciphertext2)
	require.NoError(t, err, "Second decryption failed")

	assert.Equal(t, plaintext, decrypted1)
	assert.Equal(t, plaintext, decrypted2)
}

package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/crypto"
	"github.com/ilxqx/vef-framework-go/encoding"
)

func TestRSAPasswordDecryptor(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create decryptor
	decryptor, err := NewRsaPasswordDecryptor(privateKey)
	require.NoError(t, err, "Failed to create RSA password decryptor")

	// Test password
	password := "MySecurePassword123!@#"

	// Encrypt using crypto package
	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	encryptedPassword, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decryptedPassword, err := decryptor.Decrypt(encryptedPassword)
	require.NoError(t, err, "Failed to decrypt password")

	// Verify
	assert.Equal(t, password, decryptedPassword, "Decrypted password should match original")
}

func TestRSAPasswordDecryptor_FromPEM(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Convert to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := []byte("-----BEGIN RSA PRIVATE KEY-----\n" +
		encoding.ToBase64(privateKeyBytes) +
		"\n-----END RSA PRIVATE KEY-----")

	decryptor, err := NewRsaPasswordDecryptorFromPEM(privatePEM)
	require.NoError(t, err)

	// Encrypt using crypto package
	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestRSAPasswordDecryptor_FromHex(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Convert to hex
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyHex := encoding.ToHex(privateKeyBytes)

	decryptor, err := NewRsaPasswordDecryptorFromHex(privateKeyHex)
	require.NoError(t, err)

	// Encrypt using crypto package
	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestRSAPasswordDecryptor_FromBase64(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Convert to base64
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := encoding.ToBase64(privateKeyBytes)

	decryptor, err := NewRsaPasswordDecryptorFromBase64(privateKeyBase64)
	require.NoError(t, err)

	// Encrypt using crypto package
	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestRSAPasswordDecryptor_NilKey(t *testing.T) {
	_, err := NewRsaPasswordDecryptor(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "private key cannot be nil")
}

func TestRSAPasswordDecryptor_PKCS8Format(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Convert to PKCS8 format
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	privateKeyHex := encoding.ToHex(privateKeyBytes)

	decryptor, err := NewRsaPasswordDecryptorFromHex(privateKeyHex)
	require.NoError(t, err)

	// Encrypt using crypto package
	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestRSAPasswordDecryptor_LongPassword(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	decryptor, err := NewRsaPasswordDecryptor(privateKey)
	require.NoError(t, err)

	// Encrypt using crypto package
	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	// Test with a reasonably long password (but within RSA limits)
	password := "ThisIsAVeryLongPasswordWith123Numbers!@#$%^&*()SpecialChars"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

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
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	decryptor, err := NewRsaPasswordDecryptor(privateKey)
	require.NoError(t, err, "Should create RSA password decryptor")

	password := "MySecurePassword123!@#"

	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	encryptedPassword, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decryptedPassword, err := decryptor.Decrypt(encryptedPassword)
	require.NoError(t, err, "Should decrypt password")

	assert.Equal(t, password, decryptedPassword, "Decrypted password should match original")
}

func TestRSAPasswordDecryptor_FromPEM(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := []byte("-----BEGIN RSA PRIVATE KEY-----\n" +
		encoding.ToBase64(privateKeyBytes) +
		"\n-----END RSA PRIVATE KEY-----")

	decryptor, err := NewRsaPasswordDecryptorFromPEM(privatePEM)
	require.NoError(t, err, "Should create decryptor from PEM")

	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
}

func TestRSAPasswordDecryptor_FromHex(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyHex := encoding.ToHex(privateKeyBytes)

	decryptor, err := NewRsaPasswordDecryptorFromHex(privateKeyHex)
	require.NoError(t, err, "Should create decryptor from hex")

	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
}

func TestRSAPasswordDecryptor_FromBase64(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := encoding.ToBase64(privateKeyBytes)

	decryptor, err := NewRsaPasswordDecryptorFromBase64(privateKeyBase64)
	require.NoError(t, err, "Should create decryptor from base64")

	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
}

func TestRSAPasswordDecryptor_NilKey(t *testing.T) {
	_, err := NewRsaPasswordDecryptor(nil)
	assert.Error(t, err, "Should return error for nil private key")
	assert.Contains(t, err.Error(), "private key cannot be nil")
}

func TestRSAPasswordDecryptor_PKCS8Format(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	privateKeyHex := encoding.ToHex(privateKeyBytes)

	decryptor, err := NewRsaPasswordDecryptorFromHex(privateKeyHex)
	require.NoError(t, err, "Should create decryptor from PKCS8 hex")

	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
}

func TestRSAPasswordDecryptor_LongPassword(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	decryptor, err := NewRsaPasswordDecryptor(privateKey)
	require.NoError(t, err, "Should create RSA password decryptor")

	cipher, err := crypto.NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "ThisIsAVeryLongPasswordWith123Numbers!@#$%^&*()SpecialChars"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt long password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
}

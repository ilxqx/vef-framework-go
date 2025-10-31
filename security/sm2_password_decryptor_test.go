package security

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjfoc/gmsm/sm2"

	"github.com/ilxqx/vef-framework-go/crypto"
	"github.com/ilxqx/vef-framework-go/encoding"
)

func TestSM2PasswordDecryptor(t *testing.T) {
	privateKey, err := sm2.GenerateKey(rand.Reader)
	require.NoError(t, err, "Should generate SM2 key pair")

	decryptor, err := NewSm2PasswordDecryptor(privateKey)
	require.NoError(t, err, "Should create SM2 password decryptor")

	password := "MySecurePassword123!@#"

	cipherBytes, err := sm2.Encrypt(&privateKey.PublicKey, []byte(password), rand.Reader, sm2.C1C3C2)
	require.NoError(t, err, "Should encrypt password")

	encryptedPasswordB64 := encoding.ToBase64(cipherBytes)

	decryptedPassword, err := decryptor.Decrypt(encryptedPasswordB64)
	require.NoError(t, err, "Should decrypt password")

	assert.Equal(t, password, decryptedPassword, "Decrypted password should match original")
}

func TestSM2PasswordDecryptor_NilKey(t *testing.T) {
	_, err := NewSm2PasswordDecryptor(nil)
	assert.Error(t, err, "Should return error for nil private key")
	assert.Contains(t, err.Error(), "private key cannot be nil")
}

func TestSM2PasswordDecryptor_Integration(t *testing.T) {
	privateKey, err := sm2.GenerateKey(rand.Reader)
	require.NoError(t, err, "Should generate SM2 key pair")

	decryptor, err := NewSm2PasswordDecryptor(privateKey)
	require.NoError(t, err, "Should create SM2 password decryptor")

	cipher, err := crypto.NewSM2(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	password := "TestPassword123"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
}

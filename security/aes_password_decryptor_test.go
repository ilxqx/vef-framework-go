package security

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/crypto"
	"github.com/ilxqx/vef-framework-go/encoding"
)

func TestAESPasswordDecryptor(t *testing.T) {
	key := make([]byte, 32)
	iv := make([]byte, 16)
	_, err := rand.Read(key)
	require.NoError(t, err)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	decryptor, err := NewAesPasswordDecryptor(key, iv)
	require.NoError(t, err, "Should create AES password decryptor")

	password := "MySecurePassword123!@#"

	cipher, err := crypto.NewAES(key, iv, crypto.AesModeCBC)
	require.NoError(t, err)

	encryptedPassword, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decryptedPassword, err := decryptor.Decrypt(encryptedPassword)
	require.NoError(t, err, "Should decrypt password")

	assert.Equal(t, password, decryptedPassword, "Decrypted password should match original")
}

func TestAESPasswordDecryptor_FromHex(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	ivHex := "0123456789abcdef0123456789abcdef"

	decryptor, err := NewAesPasswordDecryptorFromHex(keyHex, ivHex)
	require.NoError(t, err, "Should create decryptor from hex")

	cipher, err := crypto.NewAESFromHex(keyHex, ivHex, crypto.AesModeCBC)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
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

	decryptor, err := NewAesPasswordDecryptorFromBase64(keyBase64, ivBase64)
	require.NoError(t, err, "Should create decryptor from base64")

	cipher, err := crypto.NewAES(key, iv, crypto.AesModeCBC)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
}

func TestAESPasswordDecryptor_InvalidKeyLength(t *testing.T) {
	invalidKey := make([]byte, 15)
	iv := make([]byte, 16)

	_, err := NewAesPasswordDecryptor(invalidKey, iv)
	assert.Error(t, err, "Should return error for invalid key length")
	assert.Contains(t, err.Error(), "invalid AES key length")
}

func TestAESPasswordDecryptor_InvalidIVLength(t *testing.T) {
	key := make([]byte, 32)
	invalidIV := make([]byte, 8)

	_, err := NewAesPasswordDecryptor(key, invalidIV)
	assert.Error(t, err, "Should return error for invalid IV length")
	assert.Contains(t, err.Error(), "invalid IV length")
}

func TestAESPasswordDecryptor_NilIV(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	decryptor, err := NewAesPasswordDecryptor(key, nil)
	require.NoError(t, err, "Should create decryptor with nil IV")

	cipher, err := crypto.NewAES(key, key[:16], crypto.AesModeCBC)
	require.NoError(t, err)

	password := "TestPassword"
	encrypted, err := cipher.Encrypt(password)
	require.NoError(t, err)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err, "Should decrypt password")
	assert.Equal(t, password, decrypted, "Decrypted password should match original")
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

			decryptor, err := NewAesPasswordDecryptor(key, iv)
			require.NoError(t, err, "Should create decryptor with key size %d", tc.keySize)

			cipher, err := crypto.NewAES(key, iv, crypto.AesModeCBC)
			require.NoError(t, err)

			password := "TestPassword"
			encrypted, err := cipher.Encrypt(password)
			require.NoError(t, err)

			decrypted, err := decryptor.Decrypt(encrypted)
			require.NoError(t, err, "Should decrypt password")
			assert.Equal(t, password, decrypted, "Decrypted password should match original")
		})
	}
}

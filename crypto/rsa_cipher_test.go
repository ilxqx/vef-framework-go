package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

// TestRSACipher_OAEP tests RSA encryption and decryption in OAEP mode.
func TestRSACipher_OAEP(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey, RsaModeOAEP)
	require.NoError(t, err, "Should create RSA cipher in OAEP mode")

	tests := []struct {
		name      string
		plaintext string
	}{
		{"EnglishText", "Hello, World!"},
		{"WithDescription", "RSA-OAEP encryption test"},
		{"ChineseCharacters", "中文测试"},
		{"SpecialCharacters", "Special chars: !@#$%^&*()"},
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

// TestRSACipher_PKCS1v15 tests RSA encryption and decryption in PKCS1v15 mode.
func TestRSACipher_PKCS1v15(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey, RsaModePKCS1v15)
	require.NoError(t, err, "Should create RSA cipher in PKCS1v15 mode")

	tests := []struct {
		name      string
		plaintext string
	}{
		{"EnglishText", "Hello, World!"},
		{"WithDescription", "RSA-PKCS1v15 encryption test"},
		{"ChineseCharacters", "中文测试"},
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

// TestRSACipher_FromPEM tests creating RSA cipher from PEM-encoded keys.
func TestRSACipher_FromPEM(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Should marshal public key")

	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	cipher, err := NewRSAFromPEM(privatePEM, publicPEM, RsaModeOAEP)
	require.NoError(t, err, "Should create RSA cipher from PEM")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRSACipher_PublicKeyOnly tests RSA cipher with only public key.
func TestRSACipher_PublicKeyOnly(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(nil, &privateKey.PublicKey, RsaModeOAEP)
	require.NoError(t, err, "Should create RSA cipher with public key only")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	_, err = cipher.Decrypt(ciphertext)
	assert.Error(t, err, "Should reject decryption without private key")
}

// TestRSACipher_PrivateKeyOnly tests RSA cipher with only private key.
func TestRSACipher_PrivateKeyOnly(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(privateKey, nil, RsaModeOAEP)
	require.NoError(t, err, "Should create RSA cipher with private key only")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRSACipher_NoKeys tests that creating cipher without keys fails.
func TestRSACipher_NoKeys(t *testing.T) {
	_, err := NewRSA(nil, nil, RsaModeOAEP)
	assert.Error(t, err, "Should reject creating cipher without any keys")
}

// TestRSACipher_PKCS8PrivateKey tests creating RSA cipher from PKCS8 PEM.
func TestRSACipher_PKCS8PrivateKey(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err, "Should marshal PKCS8 private key")

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	cipher, err := NewRSAFromPEM(privatePEM, nil, RsaModeOAEP)
	require.NoError(t, err, "Should create RSA cipher from PKCS8 PEM")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRSACipher_FromHex tests creating RSA cipher from hex-encoded keys.
func TestRSACipher_FromHex(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Should marshal public key")

	publicKeyHex := hex.EncodeToString(publicKeyBytes)

	cipher, err := NewRSAFromHex(privateKeyHex, publicKeyHex, RsaModeOAEP)
	require.NoError(t, err, "Should create RSA cipher from hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRSACipher_FromHex_PKCS8 tests creating RSA cipher from PKCS8 hex-encoded key.
func TestRSACipher_FromHex_PKCS8(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err, "Should marshal PKCS8 private key")

	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	cipher, err := NewRSAFromHex(privateKeyHex, "", RsaModeOAEP)
	require.NoError(t, err, "Should create RSA cipher from PKCS8 hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRSACipher_FromBase64 tests creating RSA cipher from base64-encoded keys.
func TestRSACipher_FromBase64(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Should marshal public key")

	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	cipher, err := NewRSAFromBase64(privateKeyBase64, publicKeyBase64, RsaModeOAEP)
	require.NoError(t, err, "Should create RSA cipher from base64")

	plaintext := "Test message with base64 encoded keys"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRSACipher_DefaultMode tests that default mode is OAEP.
func TestRSACipher_DefaultMode(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err, "Should create RSA cipher with default mode")

	plaintext := "Test default mode"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")

	rsaCipher, ok := cipher.(*RsaCipher)
	require.True(t, ok)
	assert.Equal(t, RsaModeOAEP, rsaCipher.mode, "Default mode should be OAEP")
}

// TestRSACipher_KeySizes tests RSA with different key sizes.
func TestRSACipher_KeySizes(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
	}{
		{"1024-bit", 1024},
		{"2048-bit", 2048},
		{"4096-bit", 4096},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, err := generateRSAKeyPair(tt.keySize)
			require.NoError(t, err, "Should generate %d-bit RSA key", tt.keySize)

			cipher, err := NewRSA(privateKey, &privateKey.PublicKey, RsaModeOAEP)
			require.NoError(t, err, "Should create RSA cipher")

			plaintext := "Test message"
			ciphertext, err := cipher.Encrypt(plaintext)
			require.NoError(t, err, "Should encrypt plaintext successfully")

			decrypted, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err, "Should decrypt ciphertext successfully")

			assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
		})
	}
}

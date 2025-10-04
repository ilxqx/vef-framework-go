package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

func TestRSACipher_OAEP(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey, RSAModeOAEP)
	require.NoError(t, err, "Failed to create RSA cipher")

	testCases := []string{
		"Hello, World!",
		"RSA-OAEP encryption test",
		"中文测试",
		"Special chars: !@#$%^&*()",
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

func TestRSACipher_PKCS1v15(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey, RSAModePKCS1v15)
	require.NoError(t, err, "Failed to create RSA cipher")

	testCases := []string{
		"Hello, World!",
		"RSA-PKCS1v15 encryption test",
		"中文测试",
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

func TestRSACipher_FromPEM(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Failed to marshal public key")
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	cipher, err := NewRSAFromPEM(privatePEM, publicPEM, RSAModeOAEP)
	require.NoError(t, err, "Failed to create RSA cipher from PEM")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestRSACipher_PublicKeyOnly(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	// Create cipher with only public key
	cipher, err := NewRSA(nil, &privateKey.PublicKey, RSAModeOAEP)
	require.NoError(t, err, "Failed to create RSA cipher")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	// Decryption should fail (no private key)
	_, err = cipher.Decrypt(ciphertext)
	assert.Error(t, err, "Expected error for decryption without private key")
}

func TestRSACipher_PrivateKeyOnly(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	// Create cipher with only private key (public key should be derived)
	cipher, err := NewRSA(privateKey, nil, RSAModeOAEP)
	require.NoError(t, err, "Failed to create RSA cipher")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestRSACipher_NoKeys(t *testing.T) {
	_, err := NewRSA(nil, nil, RSAModeOAEP)
	assert.Error(t, err, "Expected error when creating cipher without any keys")
}

func TestRSACipher_PKCS8PrivateKey(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	// Encode private key to PKCS8 PEM
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err, "Failed to marshal PKCS8 private key")
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	cipher, err := NewRSAFromPEM(privatePEM, nil, RSAModeOAEP)
	require.NoError(t, err, "Failed to create RSA cipher from PKCS8 PEM")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestRSACipher_FromHex(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	// Encode private key to hex (PKCS1 format)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Encode public key to hex (PKIX format)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Failed to marshal public key")
	publicKeyHex := hex.EncodeToString(publicKeyBytes)

	cipher, err := NewRSAFromHex(privateKeyHex, publicKeyHex, RSAModeOAEP)
	require.NoError(t, err, "Failed to create RSA cipher from hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestRSACipher_FromHex_PKCS8(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	// Encode private key to hex (PKCS8 format)
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err, "Failed to marshal PKCS8 private key")
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	cipher, err := NewRSAFromHex(privateKeyHex, "", RSAModeOAEP)
	require.NoError(t, err, "Failed to create RSA cipher from hex (PKCS8)")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestRSACipher_FromBase64(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	// Encode private key to base64 (PKCS1 format)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)

	// Encode public key to base64 (PKIX format)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Failed to marshal public key")
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	cipher, err := NewRSAFromBase64(privateKeyBase64, publicKeyBase64, RSAModeOAEP)
	require.NoError(t, err, "Failed to create RSA cipher from base64")

	plaintext := "Test message with base64 encoded keys"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)
}

func TestRSACipher_DefaultMode(t *testing.T) {
	// Test that default mode is OAEP
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Failed to generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err, "Failed to create RSA cipher with default mode")

	plaintext := "Test default mode"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Encryption failed")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Decryption failed")

	assert.Equal(t, plaintext, decrypted)

	// Verify it's using OAEP
	rsaCipher, ok := cipher.(*RSACipher)
	require.True(t, ok)
	assert.Equal(t, RSAModeOAEP, rsaCipher.mode)
}

func TestRSACipher_KeySizes(t *testing.T) {
	keySizes := []int{1024, 2048, 4096}

	for _, size := range keySizes {
		t.Run(fmt.Sprintf("%d-bit", size), func(t *testing.T) {
			privateKey, err := generateRSAKeyPair(size)
			require.NoError(t, err, "Failed to generate %d-bit RSA key", size)

			cipher, err := NewRSA(privateKey, &privateKey.PublicKey, RSAModeOAEP)
			require.NoError(t, err, "Failed to create RSA cipher")

			plaintext := "Test message"
			ciphertext, err := cipher.Encrypt(plaintext)
			require.NoError(t, err, "Encryption failed")

			decrypted, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err, "Decryption failed")

			assert.Equal(t, plaintext, decrypted)
		})
	}
}

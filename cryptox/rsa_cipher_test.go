package cryptox

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

// TestRsaCipher_Oaep tests RSA encryption and decryption in OAEP mode.
func TestRsaCipher_Oaep(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey, WithRSAMode(RsaModeOAEP))
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

// TestRsaCipher_Pkcs1v15 tests RSA encryption and decryption in PKCS1v15 mode.
func TestRsaCipher_Pkcs1v15(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey, WithRSAMode(RsaModePKCS1v15))
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

// TestRsaCipher_FromPem tests creating RSA cipher from PEM-encoded keys.
func TestRsaCipher_FromPem(t *testing.T) {
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

	cipher, err := NewRSAFromPem(privatePEM, publicPEM, WithRSAMode(RsaModeOAEP))
	require.NoError(t, err, "Should create RSA cipher from PEM")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRsaCipher_PublicKeyOnly tests RSA cipher with only public key.
func TestRsaCipher_PublicKeyOnly(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(nil, &privateKey.PublicKey, WithRSAMode(RsaModeOAEP))
	require.NoError(t, err, "Should create RSA cipher with public key only")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	_, err = cipher.Decrypt(ciphertext)
	assert.Error(t, err, "Should reject decryption without private key")
}

// TestRsaCipher_PrivateKeyOnly tests RSA cipher with only private key.
func TestRsaCipher_PrivateKeyOnly(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(privateKey, nil, WithRSAMode(RsaModeOAEP))
	require.NoError(t, err, "Should create RSA cipher with private key only")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRsaCipher_NoKeys tests that creating cipher without keys fails.
func TestRsaCipher_NoKeys(t *testing.T) {
	_, err := NewRSA(nil, nil, WithRSAMode(RsaModeOAEP))
	assert.Error(t, err, "Should reject creating cipher without any keys")
}

// TestRsaCipher_Pkcs8PrivateKey tests creating RSA cipher from PKCS8 PEM.
func TestRsaCipher_Pkcs8PrivateKey(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err, "Should marshal PKCS8 private key")

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	cipher, err := NewRSAFromPem(privatePEM, nil, WithRSAMode(RsaModeOAEP))
	require.NoError(t, err, "Should create RSA cipher from PKCS8 PEM")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRsaCipher_FromHex tests creating RSA cipher from hex-encoded keys.
func TestRsaCipher_FromHex(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Should marshal public key")

	publicKeyHex := hex.EncodeToString(publicKeyBytes)

	cipher, err := NewRSAFromHex(privateKeyHex, publicKeyHex, WithRSAMode(RsaModeOAEP))
	require.NoError(t, err, "Should create RSA cipher from hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRsaCipher_FromHex_PKCS8 tests creating RSA cipher from PKCS8 hex-encoded key.
func TestRsaCipher_FromHex_Pkcs8(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err, "Should marshal PKCS8 private key")

	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	cipher, err := NewRSAFromHex(privateKeyHex, "", WithRSAMode(RsaModeOAEP))
	require.NoError(t, err, "Should create RSA cipher from PKCS8 hex")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRsaCipher_FromBase64 tests creating RSA cipher from base64-encoded keys.
func TestRsaCipher_FromBase64(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "Should marshal public key")

	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	cipher, err := NewRSAFromBase64(privateKeyBase64, publicKeyBase64, WithRSAMode(RsaModeOAEP))
	require.NoError(t, err, "Should create RSA cipher from base64")

	plaintext := "Test message with base64 encoded keys"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestRsaCipher_DefaultMode tests that default mode is OAEP.
func TestRsaCipher_DefaultMode(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err, "Should generate RSA key pair")

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err, "Should create RSA cipher with default mode")

	plaintext := "Test default mode"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted)
}

// TestRsaCipher_KeySizes tests RSA with different key sizes.
func TestRsaCipher_KeySizes(t *testing.T) {
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

			cipher, err := NewRSA(privateKey, &privateKey.PublicKey, WithRSAMode(RsaModeOAEP))
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

func TestRsaCipher_SignVerifyPss(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err)

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey, WithRSASignMode(RsaSignModePSS))
	require.NoError(t, err)

	tests := []struct {
		name string
		data string
	}{
		{"EnglishText", "Hello, World!"},
		{"WithDescription", "RSA-PSS signature test"},
		{"ChineseCharacters", "中文测试"},
		{"SpecialCharacters", "Special chars: !@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := cipher.Sign(tt.data)
			require.NoError(t, err)

			valid, err := cipher.Verify(tt.data, signature)
			require.NoError(t, err)
			assert.True(t, valid)

			valid, err = cipher.Verify(tt.data+"tampered", signature)
			require.NoError(t, err)
			assert.False(t, valid)
		})
	}
}

func TestRsaCipher_SignVerifyPkcs1v15(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err)

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey, WithRSASignMode(RsaSignModePKCS1v15))
	require.NoError(t, err)

	data := "Test message"
	signature, err := cipher.Sign(data)
	require.NoError(t, err)

	valid, err := cipher.Verify(data, signature)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestRsaCipher_SignWithoutPrivateKey(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err)

	cipher, err := NewRSA(nil, &privateKey.PublicKey)
	require.NoError(t, err)

	data := "Test message"
	_, err = cipher.Sign(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPrivateKeyRequiredForSign)
}

func TestRsaCipher_VerifyWithoutPublicKey(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err)

	signerCipher, err := NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	data := "Test message"
	signature, err := signerCipher.Sign(data)
	require.NoError(t, err)

	verifierCipher, err := NewRSA(privateKey, nil)
	require.NoError(t, err)

	valid, err := verifierCipher.Verify(data, signature)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestRsaCipher_InvalidSignature(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err)

	cipher, err := NewRSA(privateKey, &privateKey.PublicKey)
	require.NoError(t, err)

	data := "Test message"

	_, err = cipher.Verify(data, "invalid-base64")
	assert.Error(t, err)
}

func TestRsaCipher_DifferentSignModes(t *testing.T) {
	privateKey, err := generateRSAKeyPair(2048)
	require.NoError(t, err)

	pssCipher, err := NewRSA(privateKey, &privateKey.PublicKey, WithRSASignMode(RsaSignModePSS))
	require.NoError(t, err)

	pkcs1Cipher, err := NewRSA(privateKey, &privateKey.PublicKey, WithRSASignMode(RsaSignModePKCS1v15))
	require.NoError(t, err)

	data := "Test message"

	pssSignature, err := pssCipher.Sign(data)
	require.NoError(t, err)

	pkcs1Signature, err := pkcs1Cipher.Sign(data)
	require.NoError(t, err)

	assert.NotEqual(t, pssSignature, pkcs1Signature)

	validPss, err := pssCipher.Verify(data, pssSignature)
	require.NoError(t, err)
	assert.True(t, validPss)

	validPkcs1, err := pkcs1Cipher.Verify(data, pkcs1Signature)
	require.NoError(t, err)
	assert.True(t, validPkcs1)
}

package crypto

import (
	"crypto/rand"
	"encoding/asn1"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"

	"github.com/ilxqx/vef-framework-go/encoding"
)

func generateSM2KeyPair() (*sm2.PrivateKey, error) {
	return sm2.GenerateKey(rand.Reader)
}

// TestSM2Cipher_Encrypt_Decrypt tests SM2 encryption and decryption.
func TestSM2Cipher_Encrypt_Decrypt(t *testing.T) {
	privateKey, err := generateSM2KeyPair()
	require.NoError(t, err, "Should generate SM2 key pair")

	cipher, err := NewSM2(privateKey, &privateKey.PublicKey)
	require.NoError(t, err, "Should create SM2 cipher")

	tests := []struct {
		name      string
		plaintext string
	}{
		{"EnglishText", "Hello, World!"},
		{"WithDescription", "SM2 encryption test"},
		{"ChineseCharacters", "中文测试"},
		{"SpecialCharacters", "Special chars: !@#$%^&*()"},
		{"ChineseAlgorithm", "国密SM2加密算法"},
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

// TestSM2Cipher_FromPEM tests creating SM2 cipher from PEM-encoded keys.
func TestSM2Cipher_FromPEM(t *testing.T) {
	priv, err := generateSM2KeyPair()
	require.NoError(t, err, "Should generate SM2 key pair")

	type sm2Priv struct {
		Version       int
		PrivateKey    []byte
		NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
		PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
	}

	derPriv, err := asn1.Marshal(sm2Priv{Version: 1, PrivateKey: priv.D.Bytes()})
	require.NoError(t, err, "Should marshal SM2 private key")
	derPub, err := x509.MarshalSm2PublicKey(&priv.PublicKey)
	require.NoError(t, err, "Should marshal SM2 public key")

	pemPriv := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: derPriv})
	pemPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: derPub})

	cipher, err := NewSM2FromPEM(pemPriv, pemPub)
	require.NoError(t, err, "Should create SM2 cipher from PEM")

	plaintext := "PEM roundtrip message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")
	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")
	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestSM2Cipher_FromHex tests creating SM2 cipher from hex-encoded keys.
func TestSM2Cipher_FromHex(t *testing.T) {
	priv, err := generateSM2KeyPair()
	require.NoError(t, err, "Should generate SM2 key pair")

	type sm2Priv struct {
		Version       int
		PrivateKey    []byte
		NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
		PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
	}

	derPriv, err := asn1.Marshal(sm2Priv{Version: 1, PrivateKey: priv.D.Bytes()})
	require.NoError(t, err, "Should marshal SM2 private key")
	derPub, err := x509.MarshalSm2PublicKey(&priv.PublicKey)
	require.NoError(t, err, "Should marshal SM2 public key")

	hexPriv := encoding.ToHex(derPriv)
	hexPub := encoding.ToHex(derPub)

	cipher, err := NewSM2FromHex(hexPriv, hexPub)
	require.NoError(t, err, "Should create SM2 cipher from hex")

	plaintext := "HEX roundtrip message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")
	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")
	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestSM2Cipher_PublicKeyOnly tests SM2 cipher with only public key.
func TestSM2Cipher_PublicKeyOnly(t *testing.T) {
	privateKey, err := generateSM2KeyPair()
	require.NoError(t, err, "Should generate SM2 key pair")

	cipher, err := NewSM2(nil, &privateKey.PublicKey)
	require.NoError(t, err, "Should create SM2 cipher with public key only")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	_, err = cipher.Decrypt(ciphertext)
	assert.Error(t, err, "Should reject decryption without private key")
}

// TestSM2Cipher_PrivateKeyOnly tests SM2 cipher with only private key.
func TestSM2Cipher_PrivateKeyOnly(t *testing.T) {
	privateKey, err := generateSM2KeyPair()
	require.NoError(t, err, "Should generate SM2 key pair")

	cipher, err := NewSM2(privateKey, nil)
	require.NoError(t, err, "Should create SM2 cipher with private key only")

	plaintext := "Test message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")
}

// TestSM2Cipher_NoKeys tests that creating cipher without keys fails.
func TestSM2Cipher_NoKeys(t *testing.T) {
	_, err := NewSM2(nil, nil)
	assert.Error(t, err, "Should reject creating cipher without any keys")
}

// TestSM2Cipher_MultipleEncryptions tests that SM2 produces different ciphertexts.
func TestSM2Cipher_MultipleEncryptions(t *testing.T) {
	privateKey, err := generateSM2KeyPair()
	require.NoError(t, err, "Should generate SM2 key pair")

	cipher, err := NewSM2(privateKey, &privateKey.PublicKey)
	require.NoError(t, err, "Should create SM2 cipher")

	plaintext := "Test message"

	ciphertext1, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	ciphertext2, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "Should encrypt plaintext successfully")

	assert.NotEqual(t, ciphertext1, ciphertext2,
		"SM2 should produce different ciphertexts due to random component")

	decrypted1, err := cipher.Decrypt(ciphertext1)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	decrypted2, err := cipher.Decrypt(ciphertext2)
	require.NoError(t, err, "Should decrypt ciphertext successfully")

	assert.Equal(t, plaintext, decrypted1, "First decrypted text should match original plaintext")
	assert.Equal(t, plaintext, decrypted2, "Second decrypted text should match original plaintext")
}

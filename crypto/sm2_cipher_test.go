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
	// Generate key pair using the same library used by the implementation
	priv, err := generateSM2KeyPair()
	require.NoError(t, err, "failed to generate SM2 key pair")

	// Build raw SM2 private key DER matching x509.ParseSm2PrivateKey expectations
	type sm2Priv struct {
		Version       int
		PrivateKey    []byte
		NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
		PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
	}
	derPriv, err := asn1.Marshal(sm2Priv{Version: 1, PrivateKey: priv.D.Bytes()})
	require.NoError(t, err, "failed to marshal raw SM2 private key")
	// For public key, use library helper to ensure correct DER
	derPub, err := x509.MarshalSm2PublicKey(&priv.PublicKey)
	require.NoError(t, err, "failed to marshal SM2 public key")

	// Wrap into PEM blocks
	pemPriv := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: derPriv})
	pemPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: derPub})

	cipher, err := NewSM2FromPEM(pemPriv, pemPub)
	require.NoError(t, err, "failed to create SM2 cipher from PEM")

	// Verify encrypt/decrypt
	plaintext := "PEM roundtrip message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "encryption failed for PEM roundtrip")
	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "decryption failed for PEM roundtrip")
	assert.Equal(t, plaintext, decrypted)
}

func TestSM2Cipher_FromHex(t *testing.T) {
	// Generate key pair
	priv, err := generateSM2KeyPair()
	require.NoError(t, err, "failed to generate SM2 key pair")

	// Build raw SM2 private key DER as above and convert to HEX
	type sm2Priv struct {
		Version       int
		PrivateKey    []byte
		NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
		PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
	}
	derPriv, err := asn1.Marshal(sm2Priv{Version: 1, PrivateKey: priv.D.Bytes()})
	require.NoError(t, err, "failed to marshal raw SM2 private key")
	derPub, err := x509.MarshalSm2PublicKey(&priv.PublicKey)
	require.NoError(t, err, "failed to marshal SM2 public key")

	hexPriv := encoding.ToHex(derPriv)
	hexPub := encoding.ToHex(derPub)

	cipher, err := NewSM2FromHex(hexPriv, hexPub)
	require.NoError(t, err, "failed to create SM2 cipher from HEX")

	plaintext := "HEX roundtrip message"
	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err, "encryption failed for HEX roundtrip")
	decrypted, err := cipher.Decrypt(ciphertext)
	require.NoError(t, err, "decryption failed for HEX roundtrip")
	assert.Equal(t, plaintext, decrypted)
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

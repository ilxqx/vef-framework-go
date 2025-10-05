package crypto

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

// SM2Cipher implements Cipher interface using SM2 encryption (国密算法).
type SM2Cipher struct {
	privateKey *sm2.PrivateKey
	publicKey  *sm2.PublicKey
}

// NewSM2 creates a new SM2 cipher with the given private and public keys.
// For encryption-only operations, privateKey can be nil.
// For decryption-only operations, publicKey can be nil.
func NewSM2(privateKey *sm2.PrivateKey, publicKey *sm2.PublicKey) (Cipher, error) {
	if privateKey == nil && publicKey == nil {
		return nil, fmt.Errorf("%w", ErrAtLeastOneKeyRequired)
	}

	// If only private key is provided, derive public key
	if publicKey == nil && privateKey != nil {
		publicKey = &privateKey.PublicKey
	}

	return &SM2Cipher{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// NewSM2FromPEM creates a new SM2 cipher from PEM-encoded keys.
// Either privatePEM or publicPEM can be nil, but not both.
func NewSM2FromPEM(privatePEM, publicPEM []byte) (Cipher, error) {
	var (
		privateKey *sm2.PrivateKey
		publicKey  *sm2.PublicKey
		err        error
	)

	if privatePEM != nil {
		privateKey, err = parseSM2PrivateKeyFromPEM(privatePEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicPEM != nil {
		publicKey, err = parseSM2PublicKeyFromPEM(publicPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewSM2(privateKey, publicKey)
}

// NewSM2FromHex creates a new SM2 cipher from hex-encoded keys.
// Either privateKeyHex or publicKeyHex can be empty, but not both.
func NewSM2FromHex(privateKeyHex, publicKeyHex string) (Cipher, error) {
	var (
		privateKey *sm2.PrivateKey
		publicKey  *sm2.PublicKey
	)

	if privateKeyHex != "" {
		keyBytes, err := encoding.FromHex(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key from hex: %w", err)
		}

		privateKey, err = x509.ParseSm2PrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicKeyHex != "" {
		keyBytes, err := encoding.FromHex(publicKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key from hex: %w", err)
		}

		publicKey, err = x509.ParseSm2PublicKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewSM2(privateKey, publicKey)
}

// NewSM2FromBase64 creates a new SM2 cipher from base64-encoded keys.
// Either privateKeyBase64 or publicKeyBase64 can be empty, but not both.
func NewSM2FromBase64(privateKeyBase64, publicKeyBase64 string) (Cipher, error) {
	var (
		privateKey *sm2.PrivateKey
		publicKey  *sm2.PublicKey
	)

	if privateKeyBase64 != constants.Empty {
		keyBytes, err := encoding.FromBase64(privateKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key from base64: %w", err)
		}

		privateKey, err = x509.ParseSm2PrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicKeyBase64 != constants.Empty {
		keyBytes, err := encoding.FromBase64(publicKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key from base64: %w", err)
		}

		publicKey, err = x509.ParseSm2PublicKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewSM2(privateKey, publicKey)
}

// Encrypt encrypts the plaintext using SM2 public key and returns base64-encoded ciphertext.
func (s *SM2Cipher) Encrypt(plaintext string) (string, error) {
	if s.publicKey == nil {
		return constants.Empty, fmt.Errorf("%w", ErrPublicKeyRequiredForEncrypt)
	}

	ciphertext, err := sm2.Encrypt(s.publicKey, []byte(plaintext), rand.Reader, sm2.C1C3C2)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to encrypt: %w", err)
	}

	return encoding.ToBase64(ciphertext), nil
}

// Decrypt decrypts the base64-encoded ciphertext using SM2 private key and returns plaintext.
func (s *SM2Cipher) Decrypt(ciphertext string) (string, error) {
	if s.privateKey == nil {
		return constants.Empty, fmt.Errorf("%w", ErrPrivateKeyRequiredForDecrypt)
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := sm2.Decrypt(s.privateKey, encryptedData, sm2.C1C3C2)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// parseSM2PrivateKeyFromPEM parses SM2 private key from PEM-encoded data.
func parseSM2PrivateKeyFromPEM(pemData []byte) (*sm2.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("%w", ErrFailedDecodePEMBlock)
	}

	return x509.ParseSm2PrivateKey(block.Bytes)
}

// parseSM2PublicKeyFromPEM parses SM2 public key from PEM-encoded data.
func parseSM2PublicKeyFromPEM(pemData []byte) (*sm2.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("%w", ErrFailedDecodePEMBlock)
	}

	return x509.ParseSm2PublicKey(block.Bytes)
}
